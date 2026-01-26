package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

// ---------------------- BENCHMARK PostJSONURL ----------------------
func BenchmarkPostJSONURL(b *testing.B) {
	gin.SetMode(gin.TestMode)

	store := storage.NewInMemoryStorage()
	auditSvc := newTestAuditService()
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("userID", "benchmark-user")
		c.Next()
	})

	router.POST("/api/shorten", handler.PostJSONURL(store, "http://localhost", auditSvc))

	bodyData := map[string]string{"url": "https://example.com"}
	bodyBytes, _ := json.Marshal(bodyData)

	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ---------------------- BENCHMARK PostRawURL ----------------------
func BenchmarkPostRawURL(b *testing.B) {
	gin.SetMode(gin.TestMode)

	store := storage.NewInMemoryStorage()
	auditSvc := newTestAuditService()
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("userID", "benchmark-user")
		c.Next()
	})

	router.POST("/", handler.PostRawURL(store, "http://localhost", auditSvc))

	url := "https://example.com"
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(url)))
	req.Header.Set("Content-Type", "text/plain")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ---------------------- BENCHMARK PostBatchURL ----------------------
func BenchmarkPostBatchURL(b *testing.B) {
	gin.SetMode(gin.TestMode)

	store := storage.NewInMemoryStorage()
	router := gin.New()
	router.POST("/api/shorten/batch", handler.PostBatchURL(store, "http://localhost"))

	batch := []handler.BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://a.com"},
		{CorrelationID: "2", OriginalURL: "https://b.com"},
	}
	body, _ := json.Marshal(batch)

	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
