package service

import (
	"context"
	"github.com/ASTeterin/urlshortener/internal/model"
)

type ShortenerService interface {
	GetShortUrl(ctx context.Context, originalURL string) (*string, error)
	GetOriginalUrl(ctx context.Context, shortUrl string) (*string, error)
}

func NewShortenerService(repo model.ShortenerRepository) ShortenerService {
	return &shortenerService{
		repo: repo,
	}
}

type shortenerService struct {
	repo model.ShortenerRepository
}

func (s *shortenerService) GetShortUrl(ctx context.Context, originalURL string) (*string, error) {
	short := s.repo.Generate(ctx)
	url := model.Url{
		Short:    short,
		Original: originalURL,
	}
	err := s.repo.Store(ctx, url)
	if err != nil {
		return nil, err
	}
	return &url.Short, nil
}

func (s *shortenerService) GetOriginalUrl(ctx context.Context, shortUrl string) (*string, error) {
	url, err := s.repo.GetByShort(ctx, shortUrl)
	if err != nil {
		return nil, err
	}
	return &url.Original, nil
}
