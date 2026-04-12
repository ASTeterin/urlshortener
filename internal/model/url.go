package model

import (
	"context"
	"errors"
)

var (
	ErrUrlNotFound  = errors.New("url not found")
	ErrUrlNotStored = errors.New("url not stored")
)

const (
	ShortUrlLen = 8
	Letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Url struct {
	Uuid     int    `json:"uuid" db:"id"`
	Short    string `json:"short" db:"short_url"`
	Original string `json:"original_url" db:"original_url"`
}

type ShortenerRepository interface {
	Store(ctx context.Context, url Url) error
	GetByShort(ctx context.Context, short string) (Url, error)
	Generate(ctx context.Context) string
}
