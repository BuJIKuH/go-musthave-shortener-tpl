// Package handler содержит HTTP-хендлеры и вспомогательные функции для работы с запросами.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

// PingHandler возвращает Gin handler для проверки доступности хранилища.
//
// Выполняет "ping" к хранилищу через метод store.Ping с таймаутом 2 секунды.
// Если хранилище недоступно, возвращает HTTP 500.
// Если хранилище доступно, возвращает HTTP 200.
func PingHandler(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// проверяем доступность хранилища
		if err := store.Ping(ctx); err != nil {
			// логирование ошибки можно добавить при желании, например через zap.Logger
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	}
}
