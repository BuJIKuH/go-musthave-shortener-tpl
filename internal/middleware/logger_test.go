package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer

	writer := zapcore.AddSync(&buf)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writer, zap.InfoLevel)
	logger := zap.New(core)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Logger(logger))
	router.GET("/ping", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())

	logOutput := buf.String()
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"path":"/ping"`)
	assert.Contains(t, logOutput, `"status":200`)
	assert.Contains(t, logOutput, `"size":4`) // длина "pong"
	assert.Contains(t, logOutput, `"latency"`)
}
