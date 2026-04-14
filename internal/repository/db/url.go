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
		existingShortURL, err2 := repo.getStoredShortURL(ctx, url.Original)
		if err2 == nil {
			return &existingShortURL, model.ErrDuplicateURL
		}
		return nil, err
	}

	return &shortURL, nil
}

func (repo *URLRepository) StoreAll(ctx context.Context, urls []model.URL) ([]model.URL, error) {
	var result []model.URL
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	query := `INSERT INTO urls(short_url, original_url) VALUES ($1, $2) ON CONFLICT (short_url) DO NOTHING RETURNING short_url`
	for _, url := range urls {
		var shortURL string
		err = tx.QueryRowContext(ctx, query, url.Short, url.Original).Scan(&shortURL)

		var pgErr *pgconn.PgError
		if err != nil {
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				existingShortURL, err2 := repo.getStoredShortURL(ctx, url.Original)
				if err2 != nil {
					return nil, err2
				}
				result = append(result, model.URL{
					Short:    existingShortURL,
					Original: url.Original,
				})
				continue
			}
			tx.Rollback()
			return nil, err
		}
		result = append(result, model.URL{
			Short:    shortURL,
			Original: url.Original,
		})
	}
	return result, tx.Commit()
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

func (repo *URLRepository) ClearAll(ctx context.Context) error {
	query := `DELETE FROM urls`
	_, err := repo.db.ExecContext(ctx, query)
	return err
}

func (repo *URLRepository) getStoredShortURL(ctx context.Context, originalURL string) (string, error) {
	const checkQuery = `SELECT short_url FROM urls WHERE original_url = $1`
	var existingShortURL string
	err := repo.db.QueryRowContext(ctx, checkQuery, originalURL).Scan(&existingShortURL)
	return existingShortURL, err
}
