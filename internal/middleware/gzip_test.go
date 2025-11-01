package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func gzipCompress(t *testing.T, data []byte) *bytes.Buffer {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(data)
	assert.NoError(t, err)
	assert.NoError(t, gz.Close())
	return &buf
}

func TestGzipMiddleware(t *testing.T) {
	type args struct {
		method        string
		url           string
		body          any
		headers       map[string]string
		expectGzipOut bool
	}

	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantContent    string
	}{
		{
			name: "Decompress gzip request",
			args: args{
				method: http.MethodPost,
				url:    "/test",
				body:   map[string]string{"msg": "hello"},
				headers: map[string]string{
					"Content-Encoding": "gzip",
					"Content-Type":     "application/json",
				},
			},
			wantStatusCode: http.StatusOK,
			wantContent:    `{"msg":"hello"}`,
		},
		{
			name: "Compress gzip response",
			args: args{
				method: http.MethodGet,
				url:    "/test",
				headers: map[string]string{
					"Accept-Encoding": "gzip",
					"Content-Type":    "application/json",
				},
				expectGzipOut: true,
			},
			wantStatusCode: http.StatusOK,
			wantContent:    `"message":"compressed"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- захват логов ---
			var logBuf bytes.Buffer
			logger := zap.New(
				zapcore.NewCore(
					zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
					zapcore.AddSync(&logBuf),
					zapcore.DebugLevel,
				),
			)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(GzipMiddleware(logger))
			router.POST("/test", func(c *gin.Context) {
				data, _ := io.ReadAll(c.Request.Body)
				c.Data(http.StatusOK, "application/json", data)
			})
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "compressed"})
			})

			var reqBody io.Reader
			if tt.args.body != nil {
				raw, _ := json.Marshal(tt.args.body)
				if tt.args.headers["Content-Encoding"] == "gzip" {
					reqBody = gzipCompress(t, raw)
				} else {
					reqBody = bytes.NewReader(raw)
				}
			}

			req := httptest.NewRequest(tt.args.method, tt.args.url, reqBody)
			for k, v := range tt.args.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			bodyBytes, _ := io.ReadAll(resp.Body)

			if tt.args.expectGzipOut {
				assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
				gzr, err := gzip.NewReader(bytes.NewReader(bodyBytes))
				assert.NoError(t, err)
				defer gzr.Close()
				unzipped, _ := io.ReadAll(gzr)
				assert.Contains(t, string(unzipped), tt.wantContent)
			} else {
				assert.Contains(t, string(bodyBytes), tt.wantContent)
			}

			// Проверим, что логгер писал нужные сообщения
			logs := logBuf.String()
			if strings.Contains(tt.name, "Decompress") {
				assert.Contains(t, logs, "Decompressed incoming request")
			}
			if strings.Contains(tt.name, "Compress") {
				assert.Contains(t, logs, "Enabled gzip compression for response")
			}
		})
	}
}
