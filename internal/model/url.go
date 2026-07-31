package model

import (
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
	Store(url URL) (*string, error)
	GetByShort(short string) (URL, error)
	StoreAll(urls []URL) ([]URL, error)
	ListByUserID(userID string) ([]URL, error)
	Remove(shortURLs []string) BatchDeleteResult
	CountURLs() (int, error)
	CountUsers() (int, error)
}
