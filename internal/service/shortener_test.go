package service

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/repository/file"
)

const shortURL = "qWeRtYuI"

func Test_shortenerService_GetOriginalURL(t *testing.T) {
	ctx := context.TODO()
	var originalURL = "http://google.com"
	repo, err := file.NewURLRepository("test_get_original_url")
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

func Test_shortenerService_GetShortURL(t *testing.T) {
	var originalURL = "http://google.com"
	userID := uuid.New()
	repo, err := file.NewURLRepository("test_get_short_url")
	if err != nil {
		return
	}
	tests := []struct {
		name        string
		originalURL string
		userID      string
		wantErr     bool
		error       error
	}{
		{
			name:        "positive test",
			originalURL: originalURL,
			userID:      userID.String(),
			wantErr:     false,
			error:       nil,
		},
		{
			name:        "duplicate original url",
			originalURL: originalURL,
			userID:      userID.String(),
			wantErr:     true,
			error:       model.ErrDuplicateURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerService{
				repo: repo,
			}
			_, err := s.GetShortURL(context.TODO(), originalURL, tt.userID)
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
	userID := uuid.New()
	repo, err := file.NewURLRepository("test_list_short_url")
	if err != nil {
		return
	}
	tests := []struct {
		name            string
		originalURLsMap map[string]string
		userID          string
		wantErr         bool
		error           error
	}{
		{
			name:            "positive test",
			originalURLsMap: map[string]string{"uuid1": originalURL1, "uuid2": originalURL2},
			userID:          userID.String(),
			wantErr:         false,
			error:           nil,
		},
		{
			name:            "duplicate original urls",
			originalURLsMap: map[string]string{"uuid1": originalURL1, "uuid3": originalURL3},
			userID:          userID.String(),
			wantErr:         false,
			error:           nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerService{
				repo: repo,
			}
			_, err := s.ListShortURL(context.TODO(), tt.originalURLsMap, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_shortenerService_BatchRemove(t *testing.T) {
	ctx := context.TODO()
	var originalURL1 = "http://google.com"
	var originalURL2 = "http://yandex.ru"
	var originalURL3 = "http://test.ru"
	userID := uuid.New()
	repo, err := file.NewURLRepository("test_batch_delete")
	s := &shortenerService{
		repo:       repo,
		batchSize:  2,
		maxWorkers: 2,
	}
	shortURLMap, err := s.ListShortURL(ctx, map[string]string{"uuid1": originalURL1, "uuid2": originalURL2, "uuid3": originalURL3}, userID.String())
	if err != nil {
		return
	}
	shortURL1 := shortURLMap["uuid1"]
	shortURL2 := shortURLMap["uuid2"]

	tests := []struct {
		name        string
		urls        []string
		userID      string
		deletedURLs int
		wantErr     bool
		error       error
	}{
		{
			name:        "positive test",
			urls:        []string{shortURL1, shortURL2},
			userID:      userID.String(),
			deletedURLs: 2,
			wantErr:     false,
			error:       nil,
		},
		{
			name:        "delete already deleted URL",
			urls:        []string{shortURL1},
			userID:      userID.String(),
			deletedURLs: 0,
			wantErr:     false,
			error:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.BatchRemove(ctx, tt.urls)
			if (result.Errors != nil) != tt.wantErr {
				t.Errorf("BatchRemove() error = %v, wantErr %v", result.Errors, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result.SuccessCount, tt.deletedURLs) {
				t.Errorf("Deleted %d URLs, want %d", result.SuccessCount, tt.deletedURLs)
			}
		})
	}
}

func Test_shortenerService_GetStats(t *testing.T) {
	ctx := context.TODO()
	repo, err := file.NewURLRepository("test_get_stats")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_get_stats")

	s := &shortenerService{
		repo: repo,
	}

	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	_, _ = s.GetShortURL(ctx, "http://google.com", userID1)
	_, _ = s.GetShortURL(ctx, "http://yandex.ru", userID1)
	_, _ = s.GetShortURL(ctx, "http://test.ru", userID2)

	tests := []struct {
		name      string
		wantURLs  int
		wantUsers int
		wantErr   bool
	}{
		{
			name:      "positive test",
			wantURLs:  3,
			wantUsers: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, users, err := s.GetStats(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if urls != tt.wantURLs {
				t.Errorf("GetStats() urls = %v, want %v", urls, tt.wantURLs)
			}
			if users != tt.wantUsers {
				t.Errorf("GetStats() users = %v, want %v", users, tt.wantUsers)
			}
		})
	}
}
