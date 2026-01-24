package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHTTPObserver_Notify(t *testing.T) {
	log := zap.NewNop()
	called := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	obs := NewHTTPObserver(server.URL, log)

	err := obs.Notify(context.Background(), Event{
		Action: "test",
		URL:    "https://example.com",
	})

	assert.NoError(t, err)
	assert.True(t, called)
}
