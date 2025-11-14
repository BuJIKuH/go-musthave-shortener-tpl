package storage

import (
	"context"
	"strings"
	"sync"
)

type Storage interface {
	Save(ctx context.Context, id, url string) (string, bool, error)
	Get(id string) (string, bool)
	SaveBatch(ctx context.Context, batch map[string]string) (map[string]string, map[string]string, error)
}
type InMemoryStorage struct {
	mu              sync.RWMutex
	data            map[string]string
	originalToShort map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:            make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

func (s *InMemoryStorage) SaveBatch(ctx context.Context, batch map[string]string) (map[string]string, map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	for _, v := range batch {
		parts := strings.SplitN(v, "|", 2)
		if len(parts) != 2 {
			continue
		}

		originalURL := parts[0]
		shortID := parts[1]

		if existingID, ok := s.originalToShort[originalURL]; ok {
			conflictMap[originalURL] = existingID
			continue
		}

		s.data[shortID] = originalURL
		s.originalToShort[originalURL] = shortID
		newMap[originalURL] = shortID
	}

	return newMap, conflictMap, nil
}

func (s *InMemoryStorage) Save(ctx context.Context, id, url string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = url

	s.data[id] = url
	s.originalToShort[url] = id

	return id, false, nil
}

func (s *InMemoryStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[id]
	return url, ok
}
