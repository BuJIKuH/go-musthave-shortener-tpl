// Package storage реализует файловое хранилище URL для приложения,
// с поддержкой сохранения, пакетного сохранения, пометки удаленных URL и загрузки данных.
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
	Deleted     bool   `json:"deleted"`
}

type FileStorage struct {
	mu              sync.RWMutex
	path            string
	file            *os.File
	data            map[string]URLRecord
	originalToShort map[string]string
	userURLs        map[string][]BatchItem
	logger          *zap.Logger
	nextID          int
}

func NewFileStorage(path string, logger *zap.Logger) (*FileStorage, error) {
	fs := &FileStorage{
		path:            path,
		data:            make(map[string]URLRecord),
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

		fs.data[rec.ShortURL] = URLRecord{
			ShortID:     rec.ShortURL,
			OriginalURL: rec.OriginalURL,
			UserID:      "",
			Deleted:     rec.Deleted,
		}
		fs.originalToShort[rec.OriginalURL] = rec.ShortURL

		if rec.UUID > fs.nextID {
			fs.nextID = rec.UUID
		}
	}

	return scanner.Err()
}

func (fs *FileStorage) Save(ctx context.Context, userID, id, url string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if existing, ok := fs.originalToShort[url]; ok {
		return existing, nil
	}

	fs.nextID++
	rec := ShortURLRecord{
		UUID:        fs.nextID,
		ShortURL:    id,
		OriginalURL: url,
		Deleted:     false,
	}

	// save in-memory
	fs.data[id] = URLRecord{
		ShortID:     id,
		OriginalURL: url,
		UserID:      userID,
		Deleted:     false,
	}
	fs.originalToShort[url] = id
	fs.userURLs[userID] = append(fs.userURLs[userID], BatchItem{ShortID: id, OriginalURL: url})

	bytes, err := json.Marshal(rec)
	if err != nil {
		fs.logger.Error("Failed to marshal record", zap.Error(err))
		return "", err
	}

	if _, err := fs.file.Write(append(bytes, '\n')); err != nil {
		fs.logger.Error("Failed to append record to file", zap.Error(err))
		return "", err
	}
	return id, nil
}

func (fs *FileStorage) SaveBatch(ctx context.Context, userID string, batch []BatchItem) (map[string]string, map[string]string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	newMap := make(map[string]string)
	conflictMap := make(map[string]string)

	for _, item := range batch {
		if existing, ok := fs.originalToShort[item.OriginalURL]; ok {
			conflictMap[item.OriginalURL] = existing
			continue
		}
		fs.nextID++
		rec := ShortURLRecord{
			UUID:        fs.nextID,
			ShortURL:    item.ShortID,
			OriginalURL: item.OriginalURL,
			Deleted:     false,
		}
		fs.data[item.ShortID] = URLRecord{
			ShortID:     item.ShortID,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
			Deleted:     false,
		}
		fs.originalToShort[item.OriginalURL] = item.ShortID
		fs.userURLs[userID] = append(fs.userURLs[userID], item)

		bytes, _ := json.Marshal(rec)
		_, _ = fs.file.Write(append(bytes, '\n'))
		newMap[item.OriginalURL] = item.ShortID
	}

	return newMap, conflictMap, nil
}

func (fs *FileStorage) Get(id string) (*URLRecord, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	rec, ok := fs.data[id]
	if !ok {
		return nil, false
	}
	c := rec
	return &c, true
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

func (fs *FileStorage) MarkDeleted(userID string, shorts []string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, s := range shorts {
		rec, ok := fs.data[s]
		if !ok {
			continue
		}
		if rec.UserID != userID {
			continue
		}
		rec.Deleted = true
		fs.data[s] = rec

		fs.nextID++
		out := ShortURLRecord{
			UUID:        fs.nextID,
			ShortURL:    rec.ShortID,
			OriginalURL: rec.OriginalURL,
			Deleted:     true,
		}
		bytes, _ := json.Marshal(out)
		_, _ = fs.file.Write(append(bytes, '\n'))
	}

	return nil
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.file.Close()
}
