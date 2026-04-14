package file

import (
	"context"
	"encoding/json"
	"errors"
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
	var urlRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		b := make([]byte, model.ShortURLLen)
		for i := range b {
			b[i] = model.Letters[urlRandom.Intn(len(model.Letters))]
		}
		value := string(b)
		if _, ok := repo.storage[value]; !ok {
			return value
		}
	}
}

func (repo *urlRepository) Store(_ context.Context, url model.URL) (*string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, urlData := range repo.storage {
		if urlData.Original == url.Original {
			return &urlData.Short, model.ErrDuplicateURL
		}
	}

	url.UUID = repo.nextUUID()
	repo.storage[url.Short] = url
	return &url.Short, repo.save()
}

func (repo *urlRepository) StoreAll(ctx context.Context, urls []model.URL) ([]model.URL, error) {
	var result []model.URL
	for _, url := range urls {
		shortURL, err := repo.Store(ctx, url)
		if err != nil && !errors.Is(err, model.ErrDuplicateURL) {
			return nil, err
		}
		result = append(result, model.URL{
			Short:    *shortURL,
			Original: url.Original,
		})
	}
	return result, nil
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

func (repo *urlRepository) ClearAll(_ context.Context) error {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	repo.storage = make(map[string]model.URL)
	return repo.save()
}

func (repo *urlRepository) nextUUID() int {
	repo.uuid += 1
	return repo.uuid
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
