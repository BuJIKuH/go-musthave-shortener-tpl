package storage

import (
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileStorage struct {
	mu     sync.RWMutex
	file   string
	data   map[string]string
	logger *zap.Logger
}

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(path string, logger *zap.Logger) (*FileStorage, error) {
	fs := &FileStorage{
		file:   path,
		data:   make(map[string]string),
		logger: logger,
	}

	if _, err := os.Stat(path); err == nil {
		if err := fs.load(); err != nil {
			return nil, err
		}
	}

	logger.Info("File storage initialized", zap.String("path", path))
	return fs, nil
}

func (fs *FileStorage) Save(id, url string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data[id] = url
	if err := fs.saveToFile(); err != nil {
		fs.logger.Error("Failed to save data", zap.Error(err))
	}
}

func (fs *FileStorage) Get(id string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	url, ok := fs.data[id]
	return url, ok
}

func (fs *FileStorage) saveToFile() error {
	var records []Record
	for short, orig := range fs.data {
		records = append(records, Record{UUID: short, ShortURL: short, OriginalURL: orig})
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(fs.file, data, 0644); err != nil {
		return err
	}

	fs.logger.Info("Saved URLs to file", zap.Int("count", len(records)))
	return nil
}

func (fs *FileStorage) load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	bytes, err := os.ReadFile(fs.file)
	if err != nil {
		return err
	}

	var records []Record
	if err := json.Unmarshal(bytes, &records); err != nil {
		return err
	}

	for _, rec := range records {
		fs.data[rec.ShortURL] = rec.OriginalURL
	}

	fs.logger.Info("Loaded URLs from file", zap.Int("count", len(records)))
	return nil
}
