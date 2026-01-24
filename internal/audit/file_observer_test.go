package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestFileObserver проверяет работу FileObserver: запись событий аудита в файл.
// Создаётся временный файл, событие отправляется через Notify, после чего проверяется содержимое файла.
func TestFileObserver(t *testing.T) {
	log := zap.NewNop()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	// Создаём FileObserver с логгером-заглушкой
	obs, err := NewFileObserver(path, log)
	assert.NoError(t, err)
	defer obs.Close()

	// Отправляем тестовое событие
	err = obs.Notify(context.Background(), Event{
		Action: "delete",
		UserID: "u1",
		URL:    "https://example.com",
	})
	assert.NoError(t, err)

	// Проверяем, что событие записалось в файл
	data, _ := os.ReadFile(path)
	assert.Contains(t, string(data), `"action":"delete"`)
}
