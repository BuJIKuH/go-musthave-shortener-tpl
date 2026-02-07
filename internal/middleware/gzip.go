// Package middleware содержит Gin middleware для логирования, gzip-сжатия и авторизации.
//
// Logger — middleware для логирования HTTP-запросов с использованием zap.Logger.
// GzipMiddleware — middleware для сжатия ответа и разжатия запроса gzip.
// AuthMiddleware — middleware для авторизации.

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// gzipWriter оборачивает gin.ResponseWriter и записывает сжатые данные в Writer.
type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

// Write записывает данные в сжатый поток.
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

// GzipMiddleware возвращает Gin middleware для сжатия и разжатия HTTP-трафика.
//
// Поведение:
// 1. Если клиент прислал заголовок "Content-Encoding: gzip", middleware разжимает тело запроса.
// 2. Если клиент поддерживает "Accept-Encoding: gzip", middleware сжимает тело ответа.
// 3. Логи ошибок и операций пишутся в переданный zap.Logger.
//
// Пример использования:
//
//	r := gin.New()
//	r.Use(GzipMiddleware(logger))
func GzipMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var gr *gzip.Reader

		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				logger.Error("Failed to create gzip reader", zap.Error(err))
				c.String(http.StatusBadRequest, "invalid gzip body")
				c.Abort()
				return
			}
			gr = reader
			c.Request.Body = gr
			defer gr.Close()
			logger.Info("Decompressed incoming request",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
		}

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(c.Writer)
			c.Writer = &gzipWriter{ResponseWriter: c.Writer, Writer: gz}
			c.Header("Content-Encoding", "gzip")

			defer func() {
				if err := gz.Close(); err != nil {
					logger.Error("Failed to close gzip writer", zap.Error(err))
				}
			}()

			logger.Info("Enabled gzip compression for response",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
		}

		c.Next()
	}
}
