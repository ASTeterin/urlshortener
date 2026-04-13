package service

import (
	"context"
	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/repository/file"
	"reflect"
	"testing"
)

func Test_shortenerService_GetOriginalUrl(t *testing.T) {
	const shortUrl = "qWeRtYuI"
	var originalUrl = "http://google.com"
	ctx := context.Background()
	repo, err := file.NewURLRepository("test")
	if err != nil {
		return
	}
	url := model.URL{
		Short:    shortUrl,
		Original: originalUrl,
	}
	err = repo.Store(ctx, url)
	if err != nil {
		return
	}

	tests := []struct {
		name     string
		shortUrl string
		want     *string
		wantErr  bool
		error    error
	}{
		{
			name:     "positive test",
			shortUrl: shortUrl,
			want:     &originalUrl,
			wantErr:  false,
			error:    nil,
		},
		{
			name:     "original url not found",
			shortUrl: "123",
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
			got, err := s.GetOriginalUrl(ctx, tt.shortUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOriginalUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOriginalUrl() got = %v, want %v", got, tt.want)
			}
		})
	}
}
