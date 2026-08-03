package model

import (
	"context"
	"errors"
)

var (
	ErrURLNotFound       = errors.New("url not found")
	ErrDuplicateURL      = errors.New("duplicate url")
	ErrURLHasBeenDeleted = errors.New("url has been deleted")
)

const (
	ShortURLLen = 8
	Letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type BatchDeleteResult struct {
	SuccessCount int
	Error        error
}

type URL struct {
	UUID        int    `json:"uuid" db:"id"`
	Short       string `json:"short" db:"short_url"`
	Original    string `json:"original_url" db:"original_url"`
	CreatedBy   string `json:"created_by" db:"created_by"`
	DeletedFlag bool   `json:"is_deleted" db:"is_deleted"`
}

type ShortenerRepository interface {
	Store(ctx context.Context, url URL) (*string, error)
	GetByShort(ctx context.Context, short string) (URL, error)
	StoreAll(ctx context.Context, urls []URL) ([]URL, error)
	ListByUserID(ctx context.Context, userID string) ([]URL, error)
	Remove(ctx context.Context, shortURLs []string) BatchDeleteResult
	CountURLs(ctx context.Context) (int, error)
	CountUsers(ctx context.Context) (int, error)
}
