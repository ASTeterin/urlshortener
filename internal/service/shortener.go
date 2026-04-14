package service

import (
	"context"
	"fmt"
	"github.com/ASTeterin/urlshortener/internal/model"
)

type ShortenerService interface {
	GetShortURL(ctx context.Context, originalURL string) (*string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (*string, error)
	ListShortURL(ctx context.Context, originalURLsMap map[string]string) (map[string]string, error)
}

func NewShortenerService(repo model.ShortenerRepository) ShortenerService {
	return &shortenerService{
		repo: repo,
	}
}

type shortenerService struct {
	repo model.ShortenerRepository
}

func (s *shortenerService) GetShortURL(ctx context.Context, originalURL string) (*string, error) {
	short := s.repo.Generate(ctx)
	url := model.URL{
		Short:    short,
		Original: originalURL,
	}
	return s.repo.Store(ctx, url)
}

func (s *shortenerService) ListShortURL(ctx context.Context, originalURLsMap map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	urls := s.generateModels(ctx, originalURLsMap)
	storedURLs, err := s.repo.StoreAll(ctx, urls)
	fmt.Println("!!!!!!!!!!", storedURLs)
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

func (s *shortenerService) GetOriginalURL(ctx context.Context, shortURL string) (*string, error) {
	url, err := s.repo.GetByShort(ctx, shortURL)
	if err != nil {
		return nil, err
	}
	return &url.Original, nil
}

func (s *shortenerService) generateModels(ctx context.Context, originalURLsMap map[string]string) []model.URL {
	urls := make([]model.URL, 0, len(originalURLsMap))
	for _, v := range originalURLsMap {
		short := s.repo.Generate(ctx)
		url := model.URL{
			Short:    short,
			Original: v,
		}
		urls = append(urls, url)
	}
	return urls
}
