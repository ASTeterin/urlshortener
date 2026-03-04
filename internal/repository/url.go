package repository

import (
	"math/rand"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type urlRepository struct {
	storage map[string]model.Url
}

func NewUrlRepository() model.ShortenerRepository {
	return &urlRepository{
		storage: make(map[string]model.Url),
	}
}

func (repo *urlRepository) Generate() string {
	for {
		rand.Seed(time.Now().UnixNano())
		b := make([]byte, model.ShortUrlLen)
		for i := range b {
			b[i] = model.Letters[rand.Intn(len(model.Letters))]
		}
		value := string(b)
		if _, ok := repo.storage[value]; !ok {
			return value
		}
	}
}

func (repo *urlRepository) Store(url model.Url) {
	repo.storage[url.Short] = url
}

func (repo *urlRepository) FindByShort(short string) (model.Url, error) {
	url, ok := repo.storage[short]
	if !ok {
		return model.Url{}, model.ErrUrlNotFound
	}
	return url, nil
}
