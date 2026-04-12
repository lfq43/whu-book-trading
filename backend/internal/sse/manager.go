package sse

import "sync"

type Manager struct {
	mu      sync.RWMutex
	clients map[uint]map[chan int]struct{}
}

var DefaultManager = NewManager()

func NewManager() *Manager {
	return &Manager{
		clients: make(map[uint]map[chan int]struct{}),
	}
}

func (m *Manager) Register(userID uint) chan int {
	ch := make(chan int, 10)
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[userID]; !ok {
		m.clients[userID] = make(map[chan int]struct{})
	}
	m.clients[userID][ch] = struct{}{}
	return ch
}

func (m *Manager) Unregister(userID uint, ch chan int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if set, ok := m.clients[userID]; ok {
		delete(set, ch)
		if len(set) == 0 {
			delete(m.clients, userID)
		}
	}

	close(ch)
}

func (m *Manager) Send(userID uint, count int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if set, ok := m.clients[userID]; ok {
		for ch := range set {
			select {
			case ch <- count:
			default:
			}
		}
	}
}
