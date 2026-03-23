package repository

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type urlRepository struct {
	storage  map[string]model.Url
	filePath string
	uuid     int
	mu       sync.RWMutex
}

func NewUrlRepository(filePath string) (model.ShortenerRepository, error) {
	repo := &urlRepository{
		filePath: filePath,
		storage:  make(map[string]model.Url),
	}
	if err := repo.load(); err != nil {
		if os.IsNotExist(err) {
			if err := repo.save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return repo, nil
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

func (repo *urlRepository) Store(url model.Url) error {
	url.Uuid = repo.nextUuid()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.storage[url.Short] = url
	return repo.save()
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

func (repo *urlRepository) nextUuid() int {
	return repo.uuid + 1
}

func (repo *urlRepository) save() error {
	var values []model.Url
	for _, v := range repo.storage {
		values = append(values, v)
	}
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return os.WriteFile(repo.filePath, data, 0644)
}

func (repo *urlRepository) load() error {
	data, err := os.ReadFile(repo.filePath)
	if err != nil {
		return err
	}

	storedData := make([]model.Url, 0)
	err = json.Unmarshal(data, &storedData)
	if err != nil {
		return err
	}
	var maxUuid int
	urlToShortUrlMap := make(map[string]model.Url)
	for _, v := range storedData {
		urlToShortUrlMap[v.Short] = v
		if v.Uuid > maxUuid {
			maxUuid = v.Uuid
		}
	}

	repo.uuid = maxUuid
	repo.storage = urlToShortUrlMap
	return nil
}
