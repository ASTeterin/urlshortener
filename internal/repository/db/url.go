package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

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

func (repo *urlRepository) Store(url model.URL) (*string, error) {
	ctx := context.TODO()
	const query = `
        INSERT INTO urls (short_url, original_url, created_by)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO NOTHING
        RETURNING short_url
    `

	var shortURL string
	err := repo.db.QueryRowContext(ctx, query, url.Short, url.Original, url.CreatedBy).Scan(&shortURL)

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
	ctx := context.TODO()
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `INSERT INTO urls(short_url, original_url, created_by) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING RETURNING short_url`
	for _, url := range urls {
		var shortURL string
		err = tx.QueryRowContext(ctx, query, url.Short, url.Original, url.CreatedBy).Scan(&shortURL)

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
	ctx := context.TODO()
	query := `SELECT id, short_url, original_url, is_deleted FROM urls WHERE short_url = $1`
	url := model.URL{}
	err := repo.db.QueryRowContext(ctx, query, short).Scan(
		&url.UUID, &url.Short, &url.Original, &url.DeletedFlag)
	if errors.Is(err, sql.ErrNoRows) {
		return model.URL{}, model.ErrURLNotFound
	}
	if url.DeletedFlag {
		return url, model.ErrURLHasBeenDeleted
	}
	return url, err
}

func (repo *urlRepository) ListByUserID(userID string) ([]model.URL, error) {
	ctx := context.TODO()
	query := `SELECT id, short_url, original_url FROM urls WHERE created_by = $1`
	urls := make([]model.URL, 0)
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var url model.URL
		err = rows.Scan(&url.UUID, &url.Short, &url.Original)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func (repo *urlRepository) Remove(shortURLs []string) model.BatchDeleteResult {
	ctx := context.TODO()
	if len(shortURLs) == 0 {
		return model.BatchDeleteResult{SuccessCount: 0, Error: nil}
	}

	var filteredURLs []string
	for _, url := range shortURLs {
		if url != "" {
			filteredURLs = append(filteredURLs, url)
		}
	}

	if len(filteredURLs) == 0 {
		return model.BatchDeleteResult{SuccessCount: 0, Error: nil}
	}

	query := `
        UPDATE urls 
        SET is_deleted = TRUE
        WHERE short_url = ANY($1)
    `

	result, err := repo.db.ExecContext(ctx, query, filteredURLs)
	if err != nil {
		return model.BatchDeleteResult{SuccessCount: 0, Error: err}
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return model.BatchDeleteResult{SuccessCount: 0, Error: err}
	}

	return model.BatchDeleteResult{
		SuccessCount: int(affectedRows),
		Error:        nil,
	}
}

func (repo *urlRepository) CountURLs() (int, error) {
	ctx := context.TODO()
	query := `SELECT COUNT(*) FROM urls`
	var count int
	err := repo.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count urls query error: %w", err)
	}
	return count, nil
}

func (repo *urlRepository) CountUsers() (int, error) {
	ctx := context.TODO()
	query := `SELECT COUNT(DISTINCT created_by) FROM urls WHERE created_by IS NOT NULL`
	var count int
	err := repo.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users query error: %w", err)
	}
	return count, nil
}

func (repo *urlRepository) getStoredShortURL(originalURL string) (string, error) {
	ctx := context.TODO()
	const checkQuery = `SELECT short_url FROM urls WHERE original_url = $1`
	var existingShortURL string
	err := repo.db.QueryRowContext(ctx, checkQuery, originalURL).Scan(&existingShortURL)
	return existingShortURL, err
}
