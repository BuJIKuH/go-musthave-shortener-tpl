package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func TestFileStorage_SaveAndLoad(t *testing.T) {
	tests := []struct {
		name        string
		initialData []record
		newData     map[string]string
	}{
		{
			name: "empty file, save new records",
			newData: map[string]string{
				"abc123": "https://yandex.ru",
				"xyz789": "https://ya.ru",
			},
		},
		{
			name: "load existing file and add new Record",
			initialData: []record{
				{UUID: "1", ShortURL: "id1", OriginalURL: "https://google.com"},
			},
			newData: map[string]string{
				"id2": "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test_storage.json")

			if len(tt.initialData) > 0 {
				b, _ := json.MarshalIndent(tt.initialData, "", "  ")
				err := os.WriteFile(filePath, b, 0644)
				assert.NoError(t, err)
			}

			logger, _ := zap.NewDevelopment()
			fs, err := NewFileStorage(filePath, logger)
			assert.NoError(t, err)

			for id, url := range tt.newData {
				fs.Save(id, url)
			}

			for id, url := range tt.newData {
				got, ok := fs.Get(id)
				assert.True(t, ok)
				assert.Equal(t, url, got)
			}

			content, err := os.ReadFile(filePath)
			assert.NoError(t, err)

			var records []record
			err = json.Unmarshal(content, &records)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(records), len(tt.newData))
		})
	}
}

func TestFileStorage_PersistenceAfterRestart(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "persist.json")

	logger, _ := zap.NewDevelopment()
	fs, err := NewFileStorage(filePath, logger)
	assert.NoError(t, err)

	// Шаг 1. Сохраняем данные
	fs.Save("short1", "https://one.example")
	fs.Save("short2", "https://two.example")

	// Шаг 2. Проверяем, что они доступны
	got1, ok1 := fs.Get("short1")
	assert.True(t, ok1)
	assert.Equal(t, "https://one.example", got1)

	// Шаг 3. "Перезапуск" — создаём новый экземпляр FileStorage, который должен загрузить данные из файла
	fs2, err := NewFileStorage(filePath, logger)
	assert.NoError(t, err)

	got2, ok2 := fs2.Get("short2")
	assert.True(t, ok2)
	assert.Equal(t, "https://two.example", got2)

	// Проверяем, что количество восстановленных записей равно 2
	assert.Equal(t, 2, len(fs2.data))
}
