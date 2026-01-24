package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFileObserver(t *testing.T) {
	log := zap.NewNop()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	obs, err := NewFileObserver(path, log)
	assert.NoError(t, err)
	defer obs.Close()

	err = obs.Notify(context.Background(), Event{
		Action: "delete",
		UserID: "u1",
		URL:    "https://example.com",
	})

	assert.NoError(t, err)

	data, _ := os.ReadFile(path)
	assert.Contains(t, string(data), `"action":"delete"`)
}
