package audit

import (
	"sync"
)

type Event struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type Action string

const (
	Shorten Action = "shorten"
	Follow  Action = "follow"
)

type Receiver interface {
	Notify(event Event) error
}

type Manager struct {
	receivers []Receiver
	mu        sync.RWMutex
}

func NewAuditManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddReceiver(r Receiver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receivers = append(m.receivers, r)
}

func (m *Manager) Notify(event Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.receivers {
		go func(recv Receiver) {
			if err := recv.Notify(event); err != nil {
				// Логирование ошибки аудита (не блокирует основной поток)
			}
		}(r)
	}
}
