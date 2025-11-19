package storage

import (
	"context"
	"sync"
)

type Storage interface {
	Save(ctx context.Context, id, url string) (string, error)
	Get(id string) (string, bool)
	SaveBatch(ctx context.Context, batch []BatchItem) (map[string]string, map[string]string, error)
	Ping(ctx context.Context) error
}
type BatchItem struct {
	ShortID     string
	OriginalURL string
}

type InMemoryStorage struct {
	mu              sync.RWMutex
	data            map[string]string
	originalToShort map[string]string
}

func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:            make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

func (s *InMemoryStorage) SaveBatch(ctx context.Context, batch []BatchItem) (map[string]string, map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	for _, item := range batch {
		originalURL := item.OriginalURL
		shortID := item.ShortID

		if existing, ok := s.originalToShort[originalURL]; ok {
			conflictMap[originalURL] = existing
			continue
		}

		s.data[shortID] = originalURL
		s.originalToShort[originalURL] = shortID
		newMap[originalURL] = shortID
	}

	return newMap, conflictMap, nil
}

func (s *InMemoryStorage) Save(ctx context.Context, id, url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = url

	s.data[id] = url
	s.originalToShort[url] = id

	return id, nil
}

func (s *InMemoryStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[id]
	return url, ok
}
