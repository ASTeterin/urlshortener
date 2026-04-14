package service

import (
	"context"
	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/repository/file"
	"reflect"
	"testing"
)

func Test_shortenerService_GetOriginalURL(t *testing.T) {
	const shortURL = "qWeRtYuI"
	var originalURL = "http://google.com"
	ctx := context.Background()
	repo, err := file.NewURLRepository("test")
	if err != nil {
		return
	}
	url := model.URL{
		Short:    shortURL,
		Original: originalURL,
	}
	_, err = repo.Store(ctx, url)
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
			got, err := s.GetOriginalURL(ctx, tt.shortURL)
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
