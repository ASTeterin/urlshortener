package db

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type UrlRepository struct {
	db *sql.DB
}

func NewUrlRepository(db *sql.DB) *UrlRepository {
	repo := &UrlRepository{
		db: db,
	}
	return repo
}

func (repo *UrlRepository) Generate(ctx context.Context) string {
	rand.Seed(time.Now().UnixNano())
	for {
		b := make([]byte, model.ShortUrlLen)
		for i := range b {
			b[i] = model.Letters[rand.Intn(len(model.Letters))]
		}
		value := string(b)
		_, err := repo.GetByShort(ctx, value)
		if err != nil {
			if errors.Is(err, model.ErrUrlNotFound) {
				return value
			}
		}
	}
}

func (repo *UrlRepository) Store(ctx context.Context, url model.Url) error {
	query := `INSERT INTO urls(short_url, original_url) VALUES ($1, $2)`

	res, err := repo.db.ExecContext(ctx, query, url.Short, url.Original)
	if err != nil {
		return model.ErrUrlNotStored
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return model.ErrUrlNotStored
	}

	if rowsAffected == 0 {
		return model.ErrUrlNotStored
	}
	return nil
}

func (repo *UrlRepository) GetByShort(ctx context.Context, short string) (model.Url, error) {
	query := `SELECT id, short_url, original_url FROM urls WHERE short_url = $1`
	url := model.Url{}
	err := repo.db.QueryRowContext(ctx, query, short).Scan(
		&url.Uuid, &url.Short, &url.Original)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Url{}, model.ErrUrlNotFound
	}
	return url, err
}
