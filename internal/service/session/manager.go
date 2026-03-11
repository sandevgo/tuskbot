package session

import (
	"sync"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Manager struct {
	mu sync.Map
}

func NewManager() core.SessionManager {
	return &Manager{}
}

func (m *Manager) TryLock(sessionID string) bool {
	_, busy := m.mu.LoadOrStore(sessionID, true)
	return !busy
}

func (m *Manager) Unlock(sessionID string) {
	m.mu.Delete(sessionID)
}
