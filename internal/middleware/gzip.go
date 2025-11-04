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
	gz *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	if g.gz != nil {
		return g.gz.Write(data)
	}
	return g.ResponseWriter.Write(data)
}

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
			c.Request.Body = io.NopCloser(gr)

			logger.Info("Decompressed incoming request",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
		}

		accept := c.GetHeader("Accept-Encoding")
		shouldCompress := strings.Contains(accept, "gzip")

		gzw := &gzipWriter{ResponseWriter: c.Writer}
		c.Writer = gzw

		c.Next()

		contentType := c.Writer.Header().Get("Content-Type")
		if shouldCompress &&
			(strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/html")) {

			c.Writer.Header().Set("Content-Encoding", "gzip")
			c.Writer.Header().Del("Content-Length") // длина меняется

			gz := gzip.NewWriter(c.Writer)
			defer gz.Close()

			gzw.gz = gz

			logger.Info("Enabled gzip compression for response",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
		}

		if gr != nil {
			_ = gr.Close()
		}
	}
}
