package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func testUserMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", "test-user")
		c.Next()
	}
}

func TestDeleteUserURLs(t *testing.T) {
	store := storage.NewInMemoryStorage()

	// канал для проверки, что markFunc вызван
	callCh := make(chan service.DeleteTask, 1)

	// создаем реальный deleter с подмененной функцией markFunc
	d := service.NewDeleter(func(userID string, shorts []string) error {
		callCh <- service.DeleteTask{UserID: userID, IDs: shorts}
		return nil
	})

	router := gin.Default()
	router.Use(testUserMW())
	router.DELETE("/api/user/urls", handler.DeleteUserURLs(store, d))

	t.Run("valid request returns 202 and enqueue called", func(t *testing.T) {
		ids := []string{"abc123", "xyz789"}
		body, _ := json.Marshal(ids)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		// ждём, пока markFunc вызовется через канал, максимум 100ms
		select {
		case task := <-callCh:
			assert.Equal(t, "test-user", task.UserID)
			assert.ElementsMatch(t, ids, task.IDs)
		case <-time.After(5 * time.Second):
			t.Fatal("markFunc was not called within timeout")
		}
	})

	t.Run("empty userID returns 401", func(t *testing.T) {
		router2 := gin.Default()
		router2.DELETE("/api/user/urls", handler.DeleteUserURLs(store, d))

		ids := []string{"abc123"}
		body, _ := json.Marshal(ids)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router2.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
