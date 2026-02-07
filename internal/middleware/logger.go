// Package middleware содержит Gin middleware для логирования, gzip-сжатия и авторизации.
//
// Logger — middleware для логирования HTTP-запросов с использованием zap.Logger.
// GzipMiddleware — middleware для сжатия ответа.
// AuthMiddleware — middleware для авторизации.

package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger возвращает Gin middleware для логирования HTTP-запросов.
//
// Middleware записывает информацию о каждом запросе после его обработки:
// - HTTP метод (GET, POST и т.д.)
// - Путь запроса
// - HTTP статус ответа
// - Размер ответа в байтах
// - Задержку обработки запроса
// - IP клиента
//
// Logger используется для логирования через zap.Logger.
//
// Пример использования:
//
//	r := gin.New()
//	r.Use(Logger(logger))
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		if size < 0 {
			size = 0
		}

		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Int("size", size),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		)
	}
}
