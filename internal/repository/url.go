package repository

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type urlRepository struct {
	storage map[string]model.Url
	mu      sync.RWMutex
}

func NewUrlRepository() model.ShortenerRepository {
	return &urlRepository{
		storage: make(map[string]model.Url),
	}
}

func (repo *urlRepository) Generate() string {
	rand.Seed(time.Now().UnixNano())
	for {
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
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.storage[url.Short] = url
}

func (repo *urlRepository) GetByShort(short string) (model.Url, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	url, ok := repo.storage[short]
	if !ok {
		return model.Url{}, model.ErrUrlNotFound
	}
	return url, nil
}
