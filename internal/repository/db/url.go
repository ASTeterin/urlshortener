package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"math/rand"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	repo := &URLRepository{
		db: db,
	}
	return repo
}

func (repo *URLRepository) Generate(ctx context.Context) string {
	var urlRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		b := make([]byte, model.ShortURLLen)
		for i := range b {
			b[i] = model.Letters[urlRandom.Intn(len(model.Letters))]
		}
		value := string(b)
		_, err := repo.GetByShort(ctx, value)
		if err != nil {
			if errors.Is(err, model.ErrURLNotFound) {
				return value
			}
		}
	}
}

func (repo *URLRepository) Store(ctx context.Context, url model.URL) (*string, error) {
	const query = `
        INSERT INTO urls (short_url, original_url)
        VALUES ($1, $2)
        ON CONFLICT (short_url) DO NOTHING
        RETURNING short_url
    `

	var shortURL string
	err := repo.db.QueryRowContext(ctx, query, url.Short, url.Original).Scan(&shortURL)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		const checkQuery = `SELECT short_url FROM urls WHERE original_url = $1`
		var existingShortURL string
		err2 := repo.db.QueryRowContext(ctx, checkQuery, url.Original).Scan(&existingShortURL)

		if err2 == nil {
			return &existingShortURL, model.ErrDuplicateURL
		}
		return nil, err
	}

	return &shortURL, nil
}

func (repo *URLRepository) StoreAll(ctx context.Context, urls []model.URL) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	query := `INSERT INTO urls(short_url, original_url) VALUES ($1, $2)`
	for _, url := range urls {
		_, err = tx.ExecContext(ctx, query, url.Short, url.Original)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()

}

func (repo *URLRepository) GetByShort(ctx context.Context, short string) (model.URL, error) {
	query := `SELECT id, short_url, original_url FROM urls WHERE short_url = $1`
	url := model.URL{}
	err := repo.db.QueryRowContext(ctx, query, short).Scan(
		&url.UUID, &url.Short, &url.Original)
	if errors.Is(err, sql.ErrNoRows) {
		return model.URL{}, model.ErrURLNotFound
	}
	return url, err
}
