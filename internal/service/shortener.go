package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

const batchSize = 50

var (
	urlRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	randMu    sync.Mutex
)

type ShortenerService interface {
	GetShortURL(ctx context.Context, originalURL, userID string) (*string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (*string, error)
	ListShortURL(ctx context.Context, originalURLsMap map[string]string, userID string) (map[string]string, error)
	ListUserURLs(ctx context.Context, userID string) (map[string]string, error)
	BatchRemove(ctx context.Context, shortURLs []string) *DeleteURLResponse
	GetStats(ctx context.Context) (urls int, users int, err error)
}

func NewShortenerService(repo model.ShortenerRepository, maxWorkers int) ShortenerService {
	return &shortenerService{
		repo:       repo,
		maxWorkers: maxWorkers,
		batchSize:  batchSize,
	}
}

type shortenerService struct {
	repo       model.ShortenerRepository
	batchSize  int
	maxWorkers int
}

type DeleteURLResponse struct {
	SuccessCount int
	Errors       []error
}

func (s *shortenerService) GetShortURL(ctx context.Context, originalURL, userID string) (*string, error) {
	short := s.generateShortURL(ctx)
	url := model.URL{
		Short:     short,
		Original:  originalURL,
		CreatedBy: userID,
	}
	return s.repo.Store(ctx, url)
}

func (s *shortenerService) ListShortURL(ctx context.Context, originalURLsMap map[string]string, userID string) (map[string]string, error) {
	result := make(map[string]string, len(originalURLsMap))
	urls := s.generateModels(ctx, originalURLsMap, userID)
	storedURLs, err := s.repo.StoreAll(ctx, urls)
	if err != nil {
		return nil, err
	}
	storedURLsMap := make(map[string]string, len(storedURLs))
	for _, url := range storedURLs {
		storedURLsMap[url.Original] = url.Short
	}

	for k, v := range originalURLsMap {
		short, ok := storedURLsMap[v]
		if !ok {
			return nil, fmt.Errorf("shortener not found: %s", v)
		}
		result[k] = short
	}

	return result, nil
}

func (s *shortenerService) ListUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	storedURLs, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(storedURLs))
	for _, url := range storedURLs {
		result[url.Original] = url.Short
	}

	return result, nil
}

func (s *shortenerService) BatchRemove(ctx context.Context, shortURLs []string) *DeleteURLResponse {
	if len(shortURLs) == 0 {
		return &DeleteURLResponse{SuccessCount: 0, Errors: nil}
	}

	batches := s.splitIntoBatches(shortURLs, s.batchSize)
	if len(batches) == 0 {
		return &DeleteURLResponse{SuccessCount: 0, Errors: nil}
	}

	sem := NewSemaphore(s.maxWorkers)
	resultCh := make(chan model.BatchDeleteResult, len(batches))

	var wg sync.WaitGroup
	for _, batch := range batches {
		wg.Add(1)

		go func(batch []string) {
			defer wg.Done()

			sem.Acquire()
			defer sem.Release()

			result := s.repo.Remove(ctx, batch)
			resultCh <- result
		}(batch)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var totalSuccess int
	var allErrors []error

	for result := range resultCh {
		totalSuccess += result.SuccessCount
		if result.Error != nil {
			allErrors = append(allErrors, result.Error)
		}
	}

	return &DeleteURLResponse{
		SuccessCount: totalSuccess,
		Errors:       allErrors,
	}
}

func (s *shortenerService) GetStats(ctx context.Context) (int, int, error) {
	urls, err := s.repo.CountURLs(ctx)
	if err != nil {
		return 0, 0, err
	}
	users, err := s.repo.CountUsers(ctx)
	if err != nil {
		return 0, 0, err
	}
	return urls, users, nil
}

func (s *shortenerService) GetOriginalURL(ctx context.Context, shortURL string) (*string, error) {
	url, err := s.repo.GetByShort(ctx, shortURL)
	if err != nil {
		return nil, err
	}
	return &url.Original, nil
}

func (s *shortenerService) generateModels(ctx context.Context, originalURLsMap map[string]string, userID string) []model.URL {
	urls := make([]model.URL, 0, len(originalURLsMap))
	for _, v := range originalURLsMap {
		short := s.generateShortURL(ctx)
		url := model.URL{
			Short:     short,
			Original:  v,
			CreatedBy: userID,
		}
		urls = append(urls, url)
	}
	return urls
}

func (s *shortenerService) generateShortURL(ctx context.Context) string {
	b := make([]byte, model.ShortURLLen)
	for {
		randMu.Lock()
		for i := range b {
			b[i] = model.Letters[urlRandom.Intn(len(model.Letters))]
		}
		randMu.Unlock()

		value := string(b)
		_, err := s.repo.GetByShort(ctx, value)
		if err != nil {
			if errors.Is(err, model.ErrURLNotFound) {
				return value
			}
		}
	}
}

func (s *shortenerService) splitIntoBatches(urls []string, batchSize int) [][]string {
	expectedBatches := (len(urls) + batchSize - 1) / batchSize
	batches := make([][]string, 0, expectedBatches)

	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		batches = append(batches, urls[i:end])
	}
	return batches
}
