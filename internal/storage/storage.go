// Package storage содержит реализацию хранилищ URL.
package storage

import (
	"context"
	"sync"
)

// InMemoryStorage — потокобезопасное хранилище URL в памяти.
type InMemoryStorage struct {
	mu              sync.RWMutex
	data            map[string]URLRecord
	originalToShort map[string]string
	userURLs        map[string][]BatchItem
}

// NewInMemoryStorage создает новое in-memory хранилище.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:            make(map[string]URLRecord),
		originalToShort: make(map[string]string),
		userURLs:        make(map[string][]BatchItem),
	}
}

// Ping проверяет доступность хранилища (в памяти всегда OK).
func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

// SaveBatch сохраняет несколько URL за один вызов.
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

		rec := URLRecord{
			ShortID:     item.ShortID,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
			Deleted:     false,
		}

		s.originalToShort[item.OriginalURL] = item.ShortID
		s.data[item.ShortID] = rec
		s.userURLs[userID] = append(s.userURLs[userID], item)

		newMap[item.OriginalURL] = item.ShortID
	}

	return newMap, conflictMap, nil
}

// Save сохраняет один URL.
func (s *InMemoryStorage) Save(ctx context.Context, userID, id, url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.originalToShort[url]; ok {
		return existing, nil
	}

	rec := URLRecord{
		ShortID:     id,
		OriginalURL: url,
		UserID:      userID,
		Deleted:     false,
	}

	s.data[id] = rec
	s.originalToShort[url] = id
	s.userURLs[userID] = append(s.userURLs[userID], BatchItem{ShortID: id, OriginalURL: url})

	return id, nil
}

// Get возвращает запись URL по короткому идентификатору.
func (s *InMemoryStorage) Get(id string) (*URLRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rec, ok := s.data[id]
	if !ok {
		return nil, false
	}
	c := rec
	return &c, true
}

// GetUserURLs возвращает список URL пользователя.
func (s *InMemoryStorage) GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, ok := s.userURLs[userID]
	if !ok {
		return nil, nil
	}
	return list, nil
}

// MarkDeleted помечает список URL как удалённые для указанного пользователя.
func (s *InMemoryStorage) MarkDeleted(userID string, shorts []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, short := range shorts {
		rec, ok := s.data[short]
		if !ok || rec.UserID != userID {
			continue
		}
		rec.Deleted = true
		s.data[short] = rec
	}
	return nil
}
