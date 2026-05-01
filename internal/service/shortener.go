package service

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type ShortenerService interface {
	GetShortURL(originalURL, userID string) (*string, error)
	GetOriginalURL(shortURL string) (*string, error)
	ListShortURL(originalURLsMap map[string]string, userID string) (map[string]string, error)
	ListUserURLs(userID string) (map[string]string, error)
	BatchRemove(shortURLs []string) *DeleteURLResponse
}

func NewShortenerService(repo model.ShortenerRepository) ShortenerService {
	return &shortenerService{
		repo:       repo,
		maxWorkers: 8,
		batchSize:  50,
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

func (s *shortenerService) GetShortURL(originalURL, userID string) (*string, error) {
	short := s.generateShortURL()
	url := model.URL{
		Short:     short,
		Original:  originalURL,
		CreatedBy: userID,
	}
	return s.repo.Store(url)
}

func (s *shortenerService) ListShortURL(originalURLsMap map[string]string, userID string) (map[string]string, error) {
	result := make(map[string]string)
	urls := s.generateModels(originalURLsMap, userID)
	storedURLs, err := s.repo.StoreAll(urls)
	if err != nil {
		return nil, err
	}
	storedURLsMap := make(map[string]string)
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

func (s *shortenerService) ListUserURLs(userID string) (map[string]string, error) {
	storedURLs, err := s.repo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, url := range storedURLs {
		result[url.Original] = url.Short
	}

	return result, nil
}

func (s *shortenerService) BatchRemove(shortURLs []string) *DeleteURLResponse {
	if len(shortURLs) == 0 {
		return &DeleteURLResponse{SuccessCount: 0, Errors: nil}
	}
	batches := s.splitIntoBatches(shortURLs, s.batchSize)

	batchCh := make(chan []string, len(batches))
	resultCh := make(chan model.BatchDeleteResult, len(batches))
	doneCh := make(chan struct{})
	defer close(doneCh)

	var wg sync.WaitGroup
	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.batchWorker(doneCh, batchCh, resultCh)
		}()
	}

	go func() {
		for _, batch := range batches {
			select {
			case <-doneCh:
				return
			case batchCh <- batch:
			}
		}
		close(batchCh)
	}()

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

func (s *shortenerService) GetOriginalURL(shortURL string) (*string, error) {
	url, err := s.repo.GetByShort(shortURL)
	if err != nil {
		return nil, err
	}
	return &url.Original, nil
}

func (s *shortenerService) generateModels(originalURLsMap map[string]string, userID string) []model.URL {
	urls := make([]model.URL, 0, len(originalURLsMap))
	for _, v := range originalURLsMap {
		short := s.generateShortURL()
		url := model.URL{
			Short:     short,
			Original:  v,
			CreatedBy: userID,
		}
		urls = append(urls, url)
	}
	return urls
}

func (s *shortenerService) generateShortURL() string {
	var urlRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		b := make([]byte, model.ShortURLLen)
		for i := range b {
			b[i] = model.Letters[urlRandom.Intn(len(model.Letters))]
		}
		value := string(b)
		_, err := s.repo.GetByShort(value)
		if err != nil {
			if errors.Is(err, model.ErrURLNotFound) {
				return value
			}
		}
	}
}

func (s *shortenerService) batchWorker(doneCh chan struct{}, batchCh <-chan []string, resultCh chan<- model.BatchDeleteResult) {
	for batch := range batchCh {
		select {
		case <-doneCh:
			resultCh <- model.BatchDeleteResult{
				SuccessCount: 0,
				Error:        nil,
			}
			return
		default:
			result := s.repo.Remove(batch)
			resultCh <- result
		}
	}
}

func (s *shortenerService) splitIntoBatches(urls []string, batchSize int) [][]string {
	var batches [][]string
	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		batches = append(batches, urls[i:end])
	}
	return batches
}
