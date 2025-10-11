package db

import "sync"

type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) Save(id, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = url
}

func (s *InMemoryStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[id]
	return url, ok
}
