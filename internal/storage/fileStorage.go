package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type ShortURLRecord struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	mu              sync.RWMutex
	path            string
	file            *os.File
	data            map[string]string
	originalToShort map[string]string
	userURLs        map[string][]BatchItem
	logger          *zap.Logger
	nextID          int
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func NewFileStorage(path string, logger *zap.Logger) (*FileStorage, error) {
	fs := &FileStorage{
		path:            path,
		data:            make(map[string]string),
		originalToShort: make(map[string]string),
		userURLs:        make(map[string][]BatchItem),
		logger:          logger,
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open file storage: %w", err)
	}
	fs.file = file

	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("cannot seek file: %w", err)
	}

	if err := fs.load(); err != nil {
		logger.Warn("Failed to load storage", zap.Error(err))
	}

	fs.file.Close()
	fs.file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Warn("Failed to open file storage", zap.Error(err))
		return nil, err
	}

	logger.Info("File storage initialized",
		zap.String("path", path),
		zap.Int("count", len(fs.data)))

	return fs, nil
}

func (fs *FileStorage) SaveBatch(ctx context.Context, userID string, batch []BatchItem) (
	map[string]string, map[string]string, error,
) {

	fs.mu.Lock()
	defer fs.mu.Unlock()

	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	for _, item := range batch {
		originalURL := item.OriginalURL
		shortID := item.ShortID

		if existing, ok := fs.originalToShort[originalURL]; ok {
			conflictMap[originalURL] = existing
			continue
		}

		fs.nextID++
		fs.data[shortID] = originalURL
		fs.originalToShort[originalURL] = shortID
		newMap[originalURL] = shortID

		fs.userURLs[userID] = append(fs.userURLs[userID], item)

		rec := ShortURLRecord{
			UUID:        fs.nextID,
			ShortURL:    shortID,
			OriginalURL: originalURL,
		}

		bytes, err := json.Marshal(rec)
		if err == nil {
			_, _ = fs.file.Write(append(bytes, '\n'))
		}
	}

	return newMap, conflictMap, nil
}

func (fs *FileStorage) Save(ctx context.Context, userID, id, url string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if existing, ok := fs.originalToShort[url]; ok {
		return existing, nil
	}

	fs.nextID++
	fs.data[id] = url
	fs.originalToShort[url] = id

	fs.userURLs[userID] = append(fs.userURLs[userID], BatchItem{
		ShortID:     id,
		OriginalURL: url,
	})

	rec := ShortURLRecord{
		UUID:        fs.nextID,
		ShortURL:    id,
		OriginalURL: url,
	}

	bytes, err := json.Marshal(rec)
	if err != nil {
		fs.logger.Error("Failed to marshal record", zap.Error(err))
		return "", err
	}

	if _, err := fs.file.Write(append(bytes, '\n')); err != nil {
		fs.logger.Error("Failed to append record to file", zap.Error(err))
		return "", err
	}

	fs.logger.Info("Saved record", zap.String("short", id), zap.String("url", url))
	return id, nil
}

func (fs *FileStorage) Get(id string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	url, ok := fs.data[id]
	return url, ok
}

func (fs *FileStorage) GetUserURLs(ctx context.Context, userID string) ([]BatchItem, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	list, ok := fs.userURLs[userID]
	if !ok || len(list) == 0 {
		return nil, nil
	}

	return list, nil
}

func (fs *FileStorage) load() error {
	scanner := bufio.NewScanner(fs.file)

	for scanner.Scan() {
		var rec ShortURLRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			fs.logger.Warn("invalid record",
				zap.String("line", scanner.Text()),
				zap.Error(err))
			continue
		}

		fs.data[rec.ShortURL] = rec.OriginalURL
		fs.originalToShort[rec.OriginalURL] = rec.ShortURL

		if rec.UUID > fs.nextID {
			fs.nextID = rec.UUID
		}
	}

	return scanner.Err()
}

func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.file.Close()
}
