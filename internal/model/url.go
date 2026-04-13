package model

import (
	"context"
	"errors"
)

var (
	ErrURLNotFound  = errors.New("url not found")
	ErrURLNotStored = errors.New("url not stored")
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
	Store(ctx context.Context, url URL) error
	GetByShort(ctx context.Context, short string) (URL, error)
	Generate(ctx context.Context) string
	StoreAll(ctx context.Context, urls []URL) error
}
