package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostBatchURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	router := gin.Default()
	store := storage.NewInMemoryStorage()
	router.POST("/api/shorten/batch", handler.PostBatchURL(store, baseURL))

	t.Run("empty batch", func(t *testing.T) {
		body, _ := json.Marshal([]handler.BatchRequestItem{})
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader("invalid_json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPostRawURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	router := gin.Default()
	store := storage.NewInMemoryStorage()
	router.POST("/", handler.PostRawURL(store, baseURL))

	t.Run("valid POST", func(t *testing.T) {
		url := "https://example.com"
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, strings.HasPrefix(string(body), baseURL))
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPostJSONURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	router := gin.Default()
	store := storage.NewInMemoryStorage()
	router.POST("/api/shorten", handler.PostJSONURL(store, baseURL))

	t.Run("valid JSON", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Contains(t, string(data), baseURL)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty URL", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"url": ""})
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
