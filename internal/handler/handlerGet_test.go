package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetIDURL(t *testing.T) {
	// Общий store с одной сохранённой ссылкой
	store := storage.NewInMemoryStorage()
	store.Save(context.Background(), "sdasda", "https://practicum.yandex.ru/")

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
