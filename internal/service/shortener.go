package service

import (
	"fmt"
	"github.com/ASTeterin/urlshortener/internal/model"
)

type ShortenerService interface {
	GetShortURL(originalURL string) (*string, error)
	GetOriginalURL(shortURL string) (*string, error)
	ListShortURL(originalURLsMap map[string]string) (map[string]string, error)
}

func NewShortenerService(repo model.ShortenerRepository) ShortenerService {
	return &shortenerService{
		repo: repo,
	}
}

type shortenerService struct {
	repo model.ShortenerRepository
}

func (s *shortenerService) GetShortURL(originalURL string) (*string, error) {
	short := s.repo.Generate()
	url := model.URL{
		Short:    short,
		Original: originalURL,
	}
	return s.repo.Store(url)
}

func (s *shortenerService) ListShortURL(originalURLsMap map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	urls := s.generateModels(originalURLsMap)
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

func (s *shortenerService) GetOriginalURL(shortURL string) (*string, error) {
	url, err := s.repo.GetByShort(shortURL)
	if err != nil {
		return nil, err
	}
	return &url.Original, nil
}

func (s *shortenerService) generateModels(originalURLsMap map[string]string) []model.URL {
	urls := make([]model.URL, 0, len(originalURLsMap))
	for _, v := range originalURLsMap {
		short := s.repo.Generate()
		url := model.URL{
			Short:    short,
			Original: v,
		}
		urls = append(urls, url)
	}
	return urls
}
