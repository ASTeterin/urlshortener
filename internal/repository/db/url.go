package db

import (
	"context"
	"database/sql"
	"errors"
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

func (repo *URLRepository) Store(ctx context.Context, url model.URL) error {
	query := `INSERT INTO urls(short_url, original_url) VALUES ($1, $2)`

	res, err := repo.db.ExecContext(ctx, query, url.Short, url.Original)
	if err != nil {
		return model.ErrURLNotStored
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return model.ErrURLNotStored
	}

	if rowsAffected == 0 {
		return model.ErrURLNotStored
	}
	return nil
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
