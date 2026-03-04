package service

import (
	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/repository"
	"reflect"
	"testing"
)

func Test_shortenerService_GetOriginalUrl(t *testing.T) {
	const shortUrl = "qWeRtYuI"
	var originalUrl = "http://google.com"
	repo := repository.NewUrlRepository()
	url := model.Url{
		Short:    shortUrl,
		Original: originalUrl,
	}
	repo.Store(url)

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
			error:    model.ErrUrlNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerService{
				repo: repo,
			}
			got, err := s.GetOriginalUrl(tt.shortUrl)
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
