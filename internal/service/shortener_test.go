package service

import (
	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/repository/file"
	"reflect"
	"testing"
)

const shortURL = "qWeRtYuI"

func Test_shortenerService_GetOriginalURL(t *testing.T) {
	var originalURL = "http://google.com"
	repo, err := file.NewURLRepository("test")
	if err != nil {
		return
	}
	url := model.URL{
		Short:    shortURL,
		Original: originalURL,
	}
	_, err = repo.Store(url)
	if err != nil {
		return
	}

	tests := []struct {
		name     string
		shortURL string
		want     *string
		wantErr  bool
		error    error
	}{
		{
			name:     "positive test",
			shortURL: shortURL,
			want:     &originalURL,
			wantErr:  false,
			error:    nil,
		},
		{
			name:     "original url not found",
			shortURL: "123",
			want:     nil,
			wantErr:  true,
			error:    model.ErrURLNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerService{
				repo: repo,
			}
			got, err := s.GetOriginalURL(tt.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOriginalURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shortenerService_GetShortURL(t *testing.T) {
	var originalURL = "http://google.com"
	repo, err := file.NewURLRepository("test")
	if err != nil {
		return
	}
	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
		error       error
	}{
		{
			name:        "positive test",
			originalURL: originalURL,
			wantErr:     false,
			error:       nil,
		},
		{
			name:        "duplicate original url",
			originalURL: originalURL,
			wantErr:     true,
			error:       model.ErrDuplicateURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerService{
				repo: repo,
			}
			_, err := s.GetShortURL(originalURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_shortenerService_ListShortURL(t *testing.T) {
	var originalURL1 = "http://google.com"
	var originalURL2 = "http://yandex.ru"
	var originalURL3 = "http://test.ru"
	repo, err := file.NewURLRepository("test")
	if err != nil {
		return
	}
	tests := []struct {
		name            string
		originalURLsMap map[string]string
		wantErr         bool
		error           error
	}{
		{
			name:            "positive test",
			originalURLsMap: map[string]string{"uuid1": originalURL1, "uuid2": originalURL2},
			wantErr:         false,
			error:           nil,
		},
		{
			name:            "duplicate original urls",
			originalURLsMap: map[string]string{"uuid1": originalURL1, "uuid3": originalURL3},
			wantErr:         false,
			error:           nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerService{
				repo: repo,
			}
			_, err := s.ListShortURL(tt.originalURLsMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
