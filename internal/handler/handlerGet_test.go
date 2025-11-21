package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := storage.NewInMemoryStorage()

	// pre-fill data for a user
	userID := "user123"
	store.Save(context.Background(), userID, "abc123", "https://ya.ru")
	store.Save(context.Background(), userID, "xyz789", "https://google.com")

	router := gin.New()
	router.GET("/api/user/urls", func(c *gin.Context) {
		c.Set("userID", userID) // emulate AuthMiddleware
	}, handler.GetUserURLs(store, "http://localhost"))

	t.Run("success returns list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		expected := []map[string]string{
			{"short_url": "http://localhost/abc123", "original_url": "https://ya.ru"},
			{"short_url": "http://localhost/xyz789", "original_url": "https://google.com"},
		}

		var actual []map[string]string
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &actual))

		assert.ElementsMatch(t, expected, actual)
	})

	t.Run("no userID returns 401", func(t *testing.T) {
		router2 := gin.New()
		router2.GET("/api/user/urls", handler.GetUserURLs(store, "http://localhost"))

		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		w := httptest.NewRecorder()

		router2.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("empty list returns 204", func(t *testing.T) {
		emptyStore := storage.NewInMemoryStorage()

		router3 := gin.New()
		router3.GET("/api/user/urls", func(c *gin.Context) {
			c.Set("userID", "some-user")
		}, handler.GetUserURLs(emptyStore, "http://localhost"))

		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		w := httptest.NewRecorder()

		router3.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestGetIDURL(t *testing.T) {
	store := storage.NewInMemoryStorage()

	_, err := store.Save(context.Background(), "asdasd", "sdasda", "https://practicum.yandex.ru/")
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/:id", handler.GetIDURL(store))

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantLocation   string
	}{
		{
			name:           "valid GET redirects",
			method:         http.MethodGet,
			path:           "/sdasda",
			wantStatusCode: http.StatusTemporaryRedirect,
			wantLocation:   "https://practicum.yandex.ru/",
		},
		{
			name:           "unknown method",
			method:         http.MethodPost,
			path:           "/sdasda",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "empty path",
			method:         http.MethodGet,
			path:           "/",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "non-existent id",
			method:         http.MethodGet,
			path:           "/unknown",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code, "unexpected status code")
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
			}
		})
	}
}
