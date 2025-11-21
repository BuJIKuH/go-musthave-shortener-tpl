package storage

import (
	"context"
	"sync"
)

type Storage interface {
	Save(ctx context.Context, userID, id, url string) (string, error)
	Get(id string) (string, bool)
	SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error)
	Ping(ctx context.Context) error
	GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error)
}
type BatchItem struct {
	ShortID     string
	OriginalURL string
}

type InMemoryStorage struct {
	mu              sync.RWMutex
	data            map[string]string
	originalToShort map[string]string
	userURLs        map[string][]BatchItem
}

func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:            make(map[string]string),
		originalToShort: make(map[string]string),
		userURLs:        make(map[string][]BatchItem),
	}
}

func (s *InMemoryStorage) SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	for _, item := range batch {
		if existing, ok := s.originalToShort[item.OriginalURL]; ok {
			conflictMap[item.OriginalURL] = existing
			continue
		}

		s.originalToShort[item.OriginalURL] = item.ShortID
		s.data[item.ShortID] = item.OriginalURL

		s.userURLs[userID] = append(s.userURLs[userID], item)

		newMap[item.OriginalURL] = item.ShortID
	}

	return newMap, conflictMap, nil
}

func (s *InMemoryStorage) Save(ctx context.Context, userID, id, url string) (string, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[id] = url
	s.originalToShort[url] = id
	s.userURLs[userID] = append(s.userURLs[userID], BatchItem{ShortID: id, OriginalURL: url})

	return id, nil
}

func (s *InMemoryStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[id]
	return url, ok
}

func (s *InMemoryStorage) GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, ok := s.userURLs[userID]
	if !ok || len(list) == 0 {
		return nil, nil
	}
	return list, nil
}
