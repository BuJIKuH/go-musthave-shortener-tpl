package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/audit"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// middleware для тестов — задаёт userID
func testUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", "test-user")
		c.Next()
	}
}

// noop observer для audit
type noopObserver struct{}

func (n *noopObserver) Notify(_ context.Context, _ audit.Event) error {
	return nil
}

func newTestAuditService() *audit.Service {
	log := zap.NewNop()
	return audit.NewService(log, &noopObserver{})
}

// --- TEST POST /api/shorten/batch ---
func TestPostBatchURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseURL := "http://localhost:8080"
	router := gin.New()
	router.Use(testUser())

	store := storage.NewInMemoryStorage()
	router.POST("/api/shorten/batch", handler.PostBatchURL(store, baseURL))

	t.Run("empty batch", func(t *testing.T) {
		body, _ := json.Marshal([]handler.BatchRequestItem{})
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader("invalid_json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid batch", func(t *testing.T) {
		batch := []handler.BatchRequestItem{
			{CorrelationID: "1", OriginalURL: "https://a.com"},
			{CorrelationID: "2", OriginalURL: "https://b.com"},
		}

		body, _ := json.Marshal(batch)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var result []handler.BatchResponseItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.True(t, strings.HasPrefix(result[0].ShortURL, baseURL))
	})
}

// --- TEST POST / ---
func TestPostRawURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseURL := "http://localhost:8080"
	router := gin.New()
	router.Use(testUser())

	store := storage.NewInMemoryStorage()
	auditSvc := newTestAuditService()

	router.POST("/", handler.PostRawURL(store, baseURL, auditSvc))

	t.Run("valid POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.True(t, strings.HasPrefix(w.Body.String(), baseURL))
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- TEST POST /api/shorten ---
func TestPostJSONURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseURL := "http://localhost:8080"
	router := gin.New()
	router.Use(testUser())

	store := storage.NewInMemoryStorage()
	auditSvc := newTestAuditService()

	router.POST("/api/shorten", handler.PostJSONURL(store, baseURL, auditSvc))

	t.Run("valid JSON", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), baseURL)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty URL", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"url": ""})
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
