package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/ASTeterin/urlshortener/internal/model"
)

type mockRepo struct {
	mu   sync.RWMutex
	data map[string]model.URL // Храним только по короткой ссылке
}

func NewMockRepository() *mockRepo {
	return &mockRepo{
		data: make(map[string]model.URL, 10000),
	}
}

func (m *mockRepo) Store(u model.URL) (*string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[u.Short] = u // Одна вставка
	return &u.Short, nil
}

func (m *mockRepo) StoreAll(urls []model.URL) ([]model.URL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range urls {
		m.data[urls[i].Short] = urls[i]
	}
	return urls, nil
}

func (m *mockRepo) GetByShort(short string) (model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.data[short]; ok {
		return u, nil
	}
	return model.URL{}, model.ErrURLNotFound
}

func (m *mockRepo) ListByUserID(userID string) ([]model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]model.URL, 0, len(m.data)) // Предварительный размер
	for _, u := range m.data {
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
		if _, ok := m.data[s]; ok {
			delete(m.data, s)
			success++
		}
	}
	return model.BatchDeleteResult{SuccessCount: success, Error: nil}
}

func (m *mockRepo) CountURLs() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data), nil
}

func (m *mockRepo) CountUsers() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]struct{})
	for _, u := range m.data {
		if u.CreatedBy != "" {
			seen[u.CreatedBy] = struct{}{}
		}
	}
	return len(seen), nil
}

func newMockService() *shortenerService {
	return &shortenerService{repo: &mockRepo{data: make(map[string]model.URL)}, maxWorkers: 4, batchSize: 50}
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
