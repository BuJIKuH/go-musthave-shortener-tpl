package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

func GzipMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var gr *gzip.Reader

		// ---------- 1. Декодирование gzip-запроса ----------
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

		// ---------- 2. Подготовка gzip-ответа ----------
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

		// ---------- 3. Продолжить выполнение ----------
		c.Next()
	}
}
