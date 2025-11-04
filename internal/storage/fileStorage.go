package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type Record struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	mu     sync.RWMutex
	path   string
	file   *os.File
	data   map[string]string
	logger *zap.Logger
	nextID int
}

func NewFileStorage(path string, logger *zap.Logger) (*FileStorage, error) {
	fs := &FileStorage{
		path:   path,
		data:   make(map[string]string),
		logger: logger,
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open file storage: %w", err)
	}
	fs.file = file

	if err := fs.load(); err != nil {
		logger.Warn("Failed to load storage", zap.Error(err))
	}

	logger.Info("File storage initialized", zap.String("path", path), zap.Int("count", len(fs.data)))
	return fs, nil
}

func (fs *FileStorage) Save(id, url string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.nextID++
	fs.data[id] = url

	rec := Record{
		UUID:        fs.nextID,
		ShortURL:    id,
		OriginalURL: url,
	}

	bytes, err := json.Marshal(rec)
	if err != nil {
		fs.logger.Error("Failed to marshal record", zap.Error(err))
		return
	}

	if _, err := fs.file.Write(append(bytes, '\n')); err != nil {
		fs.logger.Error("Failed to append record to file", zap.Error(err))
	}
	fs.logger.Info("Saved record", zap.String("short", id), zap.String("url", url))
}

func (fs *FileStorage) Get(id string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	url, ok := fs.data[id]
	return url, ok
}

func (fs *FileStorage) load() error {
	scanner := bufio.NewScanner(fs.file)
	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			fs.logger.Warn("invalid record", zap.String("line", scanner.Text()), zap.Error(err))
			continue
		}
		fs.data[rec.ShortURL] = rec.OriginalURL
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
