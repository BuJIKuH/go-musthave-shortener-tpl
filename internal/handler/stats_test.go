package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- TEST GET /api/internal/stats ---
func TestGetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(testUser())

	store := storage.NewInMemoryStorage()

	// добавим данные в storage
	userID := "test-user"

	_, _ = store.Save(context.Background(), userID, "id1", "https://a.com")
	_, _ = store.Save(context.Background(), userID, "id2", "https://b.com")

	cfg := &config.Config{
		TrustedSubnet: "192.168.1.0/24",
	}

	router.GET("/api/internal/stats", handler.GetStats(store, cfg))

	t.Run("success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "192.168.1.10")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp handler.StatsResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.URLs)
		assert.Equal(t, 1, resp.Users)
	})

	t.Run("forbidden - no header", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("forbidden - ip not in subnet", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("forbidden - subnet not configured", func(t *testing.T) {

		router2 := gin.New()
		router2.Use(testUser())

		cfg := &config.Config{}

		router2.GET("/api/internal/stats", handler.GetStats(store, cfg))

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "192.168.1.10")

		w := httptest.NewRecorder()
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
