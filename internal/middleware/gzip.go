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
		var gz *gzip.Writer
		var gr *gzip.Reader
		defer func() {
			if gz != nil {
				gz.Close()
			}
			if gr != nil {
				gr.Close()
			}
		}()

		// ---------- 1. Распаковка gzip-запроса ----------
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

			logger.Info("Decompressed incoming request",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
		}

		// ---------- 2. Сжатие ответа ----------
		accept := c.GetHeader("Accept-Encoding")
		contentType := c.GetHeader("Content-Type")

		if strings.Contains(accept, "gzip") &&
			(strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/html")) {

			c.Header("Content-Encoding", "gzip")
			gz = gzip.NewWriter(c.Writer)
			c.Writer = &gzipWriter{ResponseWriter: c.Writer, Writer: gz}

			logger.Info("Enabled gzip compression for response",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
		}

		c.Next()
	}
}
