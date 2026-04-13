package file

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type urlRepository struct {
	storage  map[string]model.URL
	filePath string
	uuid     int
	mu       sync.RWMutex
}

func NewURLRepository(filePath string) (model.ShortenerRepository, error) {
	repo := &urlRepository{
		filePath: filePath,
		storage:  make(map[string]model.URL),
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

func (repo *urlRepository) Generate(_ context.Context) string {
	rand.Seed(time.Now().UnixNano())
	for {
		b := make([]byte, model.ShortURLLen)
		for i := range b {
			b[i] = model.Letters[rand.Intn(len(model.Letters))]
		}
		value := string(b)
		if _, ok := repo.storage[value]; !ok {
			return value
		}
	}
}

func (repo *urlRepository) Store(_ context.Context, url model.URL) error {
	url.UUID = repo.nextUUID()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.storage[url.Short] = url
	return repo.save()
}

func (repo *urlRepository) StoreAll(ctx context.Context, urls []model.URL) error {
	for _, url := range urls {
		if err := repo.Store(ctx, url); err != nil {
			return err
		}
	}
	return nil
}

func (repo *urlRepository) GetByShort(_ context.Context, short string) (model.URL, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	url, ok := repo.storage[short]
	if !ok {
		return model.URL{}, model.ErrURLNotFound
	}
	return url, nil
}

func (repo *urlRepository) nextUUID() int {
	return repo.uuid + 1
}

func (repo *urlRepository) save() error {
	var values []model.URL
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

	storedData := make([]model.URL, 0)
	err = json.Unmarshal(data, &storedData)
	if err != nil {
		return err
	}
	var maxUUID int
	urlToShortURLMap := make(map[string]model.URL)
	for _, v := range storedData {
		urlToShortURLMap[v.Short] = v
		if v.UUID > maxUUID {
			maxUUID = v.UUID
		}
	}

	repo.uuid = maxUUID
	repo.storage = urlToShortURLMap
	return nil
}
