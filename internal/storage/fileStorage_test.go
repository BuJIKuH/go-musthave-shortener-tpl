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
		name     string
		setup    func(fs *storage.FileStorage)
		validate func(fs *storage.FileStorage, t *testing.T)
	}{
		{
			name: "save and get single record",
			setup: func(fs *storage.FileStorage) {
				fs.Save(context.Background(), "user123", "short1", "https://ya.ru")
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				rec, ok := fs.Get("short1")
				assert.True(t, ok)
				assert.Equal(t, "https://ya.ru", rec.OriginalURL)
				assert.Equal(t, "user123", rec.UserID)
				assert.False(t, rec.Deleted)
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
						fs.Save(context.Background(), "userABC", short, url)
					}(i)
				}
				wg.Wait()
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				for i := 0; i < 10; i++ {
					short := fmt.Sprintf("short_%d", i)
					rec, ok := fs.Get(short)
					assert.True(t, ok)
					assert.Equal(t, fmt.Sprintf("https://example.com/%d", i), rec.OriginalURL)
					assert.Equal(t, "userABC", rec.UserID)
					assert.False(t, rec.Deleted)
				}
			},
		},
		{
			name: "MarkDeleted updates records",
			setup: func(fs *storage.FileStorage) {
				fs.Save(context.Background(), "userDel", "shortDel1", "https://del1.com")
				fs.Save(context.Background(), "userDel", "shortDel2", "https://del2.com")
				fs.MarkDeleted("userDel", []string{"shortDel1"})
			},
			validate: func(fs *storage.FileStorage, t *testing.T) {
				rec1, ok1 := fs.Get("shortDel1")
				rec2, ok2 := fs.Get("shortDel2")
				assert.True(t, ok1)
				assert.True(t, ok2)
				assert.True(t, rec1.Deleted)
				assert.False(t, rec2.Deleted)
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

		fs.Save(context.Background(), "UID", "shortX", "https://jsonl-format.ru")

		file, _ := os.Open(filePath)
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCount := 0

		for scanner.Scan() {
			var rec storage.ShortURLRecord
			err := json.Unmarshal(scanner.Bytes(), &rec)
			assert.NoError(t, err)
			assert.NotEmpty(t, rec.ShortURL)
			assert.NotEmpty(t, rec.OriginalURL)
			lineCount++
		}

		assert.GreaterOrEqual(t, lineCount, 1)
	})
}

func TestFileStorage_BatchAndDeleted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	filePath := filepath.Join(t.TempDir(), "batch_test.jsonl")

	fs, err := storage.NewFileStorage(filePath, logger)
	assert.NoError(t, err)
	defer fs.Close()

	ctx := context.Background()
	userID := "user1"

	t.Run("save batch", func(t *testing.T) {
		batch := []storage.BatchItem{
			{ShortID: "s1", OriginalURL: "https://a.com"},
			{ShortID: "s2", OriginalURL: "https://b.com"},
		}

		newMap, conflictMap, err := fs.SaveBatch(ctx, userID, batch)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"https://a.com": "s1", "https://b.com": "s2"}, newMap)
		assert.Empty(t, conflictMap)

		urls, _ := fs.GetUserURLs(ctx, userID)
		assert.Len(t, urls, 2)
	})

	t.Run("mark deleted", func(t *testing.T) {
		err := fs.MarkDeleted(userID, []string{"s1"})
		assert.NoError(t, err)

		rec, ok := fs.Get("s1")
		assert.True(t, ok)
		assert.True(t, rec.Deleted)

		rec2, ok2 := fs.Get("s2")
		assert.True(t, ok2)
		assert.False(t, rec2.Deleted)
	})

	t.Run("reload from file preserves deleted", func(t *testing.T) {
		// закроем и создадим новый FileStorage
		fs.Close()
		fs2, err := storage.NewFileStorage(filePath, logger)
		assert.NoError(t, err)
		defer fs2.Close()

		rec, ok := fs2.Get("s1")
		assert.True(t, ok)
		assert.True(t, rec.Deleted)

		rec2, ok2 := fs2.Get("s2")
		assert.True(t, ok2)
		assert.False(t, rec2.Deleted)
	})
}
