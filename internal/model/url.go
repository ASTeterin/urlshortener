package model

import "errors"

var ErrUrlNotFound = errors.New("url not found")

const (
	ShortUrlLen = 8
	Letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Url struct {
	Uuid     int    `json:"uuid"`
	Short    string `json:"short"`
	Original string `json:"original_url"`
}

type ShortenerRepository interface {
	Store(url Url) error
	GetByShort(short string) (Url, error)
	Generate() string
	NextUuid() int
}
