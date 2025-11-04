package storage_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFileStorage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_storage.jsonl")

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
				fs.Save("short1", "https://ya.ru")
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				url, ok := fs.Get("short1")
				assert.True(t, ok)
				assert.Equal(t, "https://ya.ru", url)
			},
			description: "Проверяем, что запись сохраняется и доступна через Get",
		},
		{
			name: "data persists after reload",
			setup: func(fs *storage.FileStorage) {
				fs.Save("short2", "https://google.com")
			},
			validate: func(_ *storage.FileStorage, t *testing.T) {
				newLogger, _ := zap.NewDevelopment()
				newFS, err := storage.NewFileStorage(filePath, newLogger)
				assert.NoError(t, err)
				defer newFS.Close()

				url, ok := newFS.Get("short2")
				assert.True(t, ok)
				assert.Equal(t, "https://google.com", url)
			},
			description: "Проверяем персистентность между сессиями",
		},
		{
			name: "concurrent saves are safe",
			setup: func(fs *storage.FileStorage) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						fs.Save(
							"short"+string(rune('A'+i)),
							"https://example.com/"+string(rune('A'+i)),
						)
					}(i)
				}
				wg.Wait()
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				for i := 0; i < 10; i++ {
					url, ok := fs.Get("short" + string(rune('A'+i)))
					assert.True(t, ok)
					assert.Contains(t, url, "https://example.com/")
				}
			},
			description: "Проверяем, что одновременные Save() не ломают данные",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := storage.NewFileStorage(filePath, logger)
			assert.NoError(t, err)
			defer fs.Close()

			tt.setup(fs)
			tt.validate(fs, t)
		})
	}

	t.Run("file format is JSONL", func(t *testing.T) {
		fs, err := storage.NewFileStorage(filePath, logger)
		assert.NoError(t, err)
		defer fs.Close()

		fs.Save("shortX", "https://jsonl-format.ru")

		file, _ := os.Open(filePath)
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCount := 0
		for scanner.Scan() {
			var rec storage.Record
			err := json.Unmarshal(scanner.Bytes(), &rec)
			assert.NoError(t, err)
			assert.NotEmpty(t, rec.ShortURL)
			lineCount++
		}
		assert.GreaterOrEqual(t, lineCount, 1, "в файле должны быть строки JSONL")
	})
}
