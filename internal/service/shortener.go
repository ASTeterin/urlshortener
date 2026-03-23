package service

import (
	"github.com/ASTeterin/urlshortener/internal/model"
)

type ShortenerService interface {
	GetShortUrl(originalURL string) (*string, error)
	GetOriginalUrl(shortUrl string) (*string, error)
}

func NewShortenerService(repo model.ShortenerRepository) ShortenerService {
	return &shortenerService{
		repo: repo,
	}
}

type shortenerService struct {
	repo model.ShortenerRepository
}

func (s *shortenerService) GetShortUrl(originalURL string) (*string, error) {
	short := s.repo.Generate()
	url := model.Url{
		Short:    short,
		Original: originalURL,
	}
	err := s.repo.Store(url)
	if err != nil {
		return nil, err
	}
	return &url.Short, nil
}

func (s *shortenerService) GetOriginalUrl(shortUrl string) (*string, error) {
	url, err := s.repo.GetByShort(shortUrl)
	if err != nil {
		return nil, err
	}
	return &url.Original, nil
}
