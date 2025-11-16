package storage_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFileStorage(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		setup       func(fs *storage.FileStorage)
		validate    func(fs *storage.FileStorage, t *testing.T)
		description string
	}{
		{
			name: "save and get single record",
			setup: func(fs *storage.FileStorage) {
				fs.Save(context.Background(), "short1", "https://ya.ru")
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				url, ok := fs.Get("short1")
				assert.True(t, ok)
				assert.Equal(t, "https://ya.ru", url)
			},
		},
		{
			name: "concurrent saves are safe",
			setup: func(fs *storage.FileStorage) {
				var wg sync.WaitGroup
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						short := fmt.Sprintf("short_%d", i)
						url := fmt.Sprintf("https://example.com/%d", i)
						fs.Save(context.Background(), short, url)
					}(i)
				}
				wg.Wait()
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				for i := 0; i < 10; i++ {
					short := fmt.Sprintf("short_%d", i)
					url, ok := fs.Get(short)
					assert.True(t, ok)
					assert.Equal(t, fmt.Sprintf("https://example.com/%d", i), url)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			filePath := filepath.Join(t.TempDir(), "test_storage.jsonl")

			fs, err := storage.NewFileStorage(filePath, logger)
			assert.NoError(t, err)
			defer fs.Close()

			tt.setup(fs)
			tt.validate(fs, t)
		})
	}

	t.Run("file format is JSONL", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "test_jsonl.jsonl")

		fs, err := storage.NewFileStorage(filePath, logger)
		assert.NoError(t, err)
		defer fs.Close()

		fs.Save(context.Background(), "shortX", "https://jsonl-format.ru")

		file, _ := os.Open(filePath)
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCount := 0

		for scanner.Scan() {
			var rec storage.ShortURLRecord
			err := json.Unmarshal(scanner.Bytes(), &rec)
			assert.NoError(t, err)
			assert.NotEmpty(t, rec.ShortURL)
			lineCount++
		}

		assert.GreaterOrEqual(t, lineCount, 1)
	})
}
