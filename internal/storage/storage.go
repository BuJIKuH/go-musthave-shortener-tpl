package storage

import (
	"context"
	"sync"
)

type Storage interface {
	Save(id, url string)
	Get(id string) (string, bool)
	SaveBatch(ctx context.Context, batch map[string]string) error
}
type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) SaveBatch(ctx context.Context, batch map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, url := range batch {
		s.data[id] = url
	}
	return nil
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
