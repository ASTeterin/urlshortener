package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"

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
	const query = `
        INSERT INTO urls (short_url, original_url, created_by)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO NOTHING
        RETURNING short_url
    `

	var shortURL string
	err := repo.db.QueryRow(query, url.Short, url.Original, url.CreatedBy).Scan(&shortURL)

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
	defer tx.Rollback()
	query := `INSERT INTO urls(short_url, original_url, created_by) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING RETURNING short_url`
	for _, url := range urls {
		var shortURL string
		err = tx.QueryRow(query, url.Short, url.Original, url.CreatedBy).Scan(&shortURL)

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
	query := `SELECT id, short_url, original_url, is_deleted FROM urls WHERE short_url = $1`
	url := model.URL{}
	err := repo.db.QueryRow(query, short).Scan(
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
	query := `SELECT id, short_url, original_url FROM urls WHERE created_by = $1`
	urls := make([]model.URL, 0)
	rows, err := repo.db.Query(query, userID)
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
	return urls, nil
}

func (repo *urlRepository) Remove(shortURLs []string, userID string) model.BatchDeleteResult {
	if len(shortURLs) == 0 {
		return model.BatchDeleteResult{
			SuccessCount: 0,
			Error:        nil,
		}
	}
	placeholders := make([]string, len(shortURLs))
	for i := range shortURLs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
        UPDATE urls 
        SET is_deleted = 1
        WHERE short_url IN (%s) AND created_by = $%d
    `, strings.Join(placeholders, ", "), len(shortURLs)+1)

	args := []interface{}{}
	for _, url := range shortURLs {
		args = append(args, url)
	}
	args = append(args, userID)

	result, err := repo.db.Exec(query, args...)
	if err != nil {
		return model.BatchDeleteResult{
			SuccessCount: 0,
			Error:        err,
		}
	}
	affectedRows, err := result.RowsAffected()
	return model.BatchDeleteResult{
		SuccessCount: int(affectedRows),
		Error:        err,
	}
}

func (repo *urlRepository) getStoredShortURL(originalURL string) (string, error) {
	const checkQuery = `SELECT short_url FROM urls WHERE original_url = $1`
	var existingShortURL string
	err := repo.db.QueryRow(checkQuery, originalURL).Scan(&existingShortURL)
	return existingShortURL, err
}
