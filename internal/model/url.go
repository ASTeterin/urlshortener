package model

import "errors"

var ErrUrlNotFound = errors.New("url not found")

const (
	ShortUrlLen = 8
	Letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Url struct {
	Short    string
	Original string
}
