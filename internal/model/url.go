package model

import (
	"errors"
)

var (
	ErrURLNotFound  = errors.New("url not found")
	ErrDuplicateURL = errors.New("duplicate url")
)

const (
	ShortURLLen = 8
	Letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type URL struct {
	UUID     int    `json:"uuid" db:"id"`
	Short    string `json:"short" db:"short_url"`
	Original string `json:"original_url" db:"original_url"`
}

type ShortenerRepository interface {
	Store(url URL) (*string, error)
	GetByShort(short string) (URL, error)
	StoreAll(urls []URL) ([]URL, error)
	ClearAll() error
}
