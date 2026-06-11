package service

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"testing"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type mockRepo struct {
	mu    sync.RWMutex
	urls  map[string]model.URL
	short map[string]model.URL
}

func (m *mockRepo) Store(u model.URL) (*string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.urls[u.Original] = u
	m.short[u.Short] = u
	return &u.Short, nil
}
func (m *mockRepo) StoreAll(urls []model.URL) ([]model.URL, error) {
	for _, u := range urls {
		m.Store(u)
	}
	return urls, nil
}
func (m *mockRepo) GetByShort(short string) (model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.short[short]
	if !ok {
		return model.URL{}, model.ErrURLNotFound
	}
	return u, nil
}
func (m *mockRepo) ListByUserID(userID string) ([]model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []model.URL
	for _, u := range m.urls {
		if u.CreatedBy == userID {
			res = append(res, u)
		}
	}
	return res, nil
}
func (m *mockRepo) Remove(shorts []string) model.BatchDeleteResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	var success int
	for _, s := range shorts {
		if _, ok := m.short[s]; ok {
			delete(m.short, s)
			delete(m.urls, m.short[s].Original) // упрощённо
			success++
		}
	}
	return model.BatchDeleteResult{SuccessCount: success, Error: nil}
}

func newMockService() *shortenerService {
	return &shortenerService{repo: &mockRepo{urls: make(map[string]model.URL), short: make(map[string]model.URL)}, maxWorkers: 4, batchSize: 50}
}

func BenchmarkGetShortURL(b *testing.B) {
	svc := newMockService()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetShortURL("https://example.com/page", "user1")
	}
}

func BenchmarkGetOriginalURL(b *testing.B) {
	svc := newMockService()
	short, _ := svc.GetShortURL("https://example.com/target", "user1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetOriginalURL(*short)
	}
}

func BenchmarkBatchRemove(b *testing.B) {
	svc := newMockService()
	var shorts []string
	for i := 0; i < 1000; i++ {
		s, _ := svc.GetShortURL("https://example.com/x", "user1")
		shorts = append(shorts, *s)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			svc.BatchRemove(shorts)
		}
	})
}

func BenchmarkListShortURL(b *testing.B) {
	svc := newMockService()
	originalURLsMap := make(map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		u := uuid.New()
		url := fmt.Sprintf("https://example.com/%s", u.String())
		originalURLsMap[u.String()] = url
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ListShortURL(originalURLsMap, "user1")
	}
}

func BenchmarkListUserURLs(b *testing.B) {
	svc := newMockService()
	for i := 0; i < 1000; i++ {
		svc.GetShortURL("https://example.com/x", "user1")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = svc.ListUserURLs("user1")
	}
}
