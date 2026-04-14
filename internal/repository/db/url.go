package db

import (
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"math/rand"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type urlRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) model.ShortenerRepository {
	repo := &urlRepository{
		db: db,
	}
	return repo
}

func (repo *urlRepository) Generate() string {
	var urlRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		b := make([]byte, model.ShortURLLen)
		for i := range b {
			b[i] = model.Letters[urlRandom.Intn(len(model.Letters))]
		}
		value := string(b)
		_, err := repo.GetByShort(value)
		if err != nil {
			if errors.Is(err, model.ErrURLNotFound) {
				return value
			}
		}
	}
}

func (repo *urlRepository) Store(url model.URL) (*string, error) {
	const query = `
        INSERT INTO urls (short_url, original_url)
        VALUES ($1, $2)
        ON CONFLICT (short_url) DO NOTHING
        RETURNING short_url
    `

	var shortURL string
	err := repo.db.QueryRow(query, url.Short, url.Original).Scan(&shortURL)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		existingShortURL, err2 := repo.getStoredShortURL(url.Original)
		if err2 == nil {
			return &existingShortURL, model.ErrDuplicateURL
		}
		return nil, err
	}

	return &shortURL, nil
}

func (repo *urlRepository) StoreAll(urls []model.URL) ([]model.URL, error) {
	var result []model.URL
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	query := `INSERT INTO urls(short_url, original_url) VALUES ($1, $2) ON CONFLICT (short_url) DO NOTHING RETURNING short_url`
	for _, url := range urls {
		var shortURL string
		err = tx.QueryRow(query, url.Short, url.Original).Scan(&shortURL)

		var pgErr *pgconn.PgError
		if err != nil {
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				existingShortURL, err2 := repo.getStoredShortURL(url.Original)
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

func (repo *urlRepository) GetByShort(short string) (model.URL, error) {
	query := `SELECT id, short_url, original_url FROM urls WHERE short_url = $1`
	url := model.URL{}
	err := repo.db.QueryRow(query, short).Scan(
		&url.UUID, &url.Short, &url.Original)
	if errors.Is(err, sql.ErrNoRows) {
		return model.URL{}, model.ErrURLNotFound
	}
	return url, err
}

func (repo *urlRepository) ClearAll() error {
	query := `DELETE FROM urls`
	_, err := repo.db.Exec(query)
	return err
}

func (repo *urlRepository) getStoredShortURL(originalURL string) (string, error) {
	const checkQuery = `SELECT short_url FROM urls WHERE original_url = $1`
	var existingShortURL string
	err := repo.db.QueryRow(checkQuery, originalURL).Scan(&existingShortURL)
	return existingShortURL, err
}
