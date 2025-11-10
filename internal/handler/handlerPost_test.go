package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostBatchURL(t *testing.T) {
	type batchItem struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	tests := []struct {
		name           string
		method         string
		body           any
		wantStatusCode int
		wantContent    string
	}{
		{
			name:   "Valid batch request",
			method: http.MethodPost,
			body: []batchItem{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
				{CorrelationID: "2", OriginalURL: "https://openai.com"},
			},
			wantStatusCode: http.StatusCreated,
			wantContent:    `"short_url":"http://localhost:8080/`,
		},
		{
			name:           "Empty array",
			method:         http.MethodPost,
			body:           []batchItem{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           `invalid_json`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Wrong HTTP method",
			method:         http.MethodGet,
			body:           nil,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewInMemoryStorage()
			router := gin.Default()
			router.POST("/api/shorten/batch", PostBatchURL(store, "http://localhost:8080"))

			var reqBody *bytes.Reader
			switch b := tt.body.(type) {
			case string:
				reqBody = bytes.NewReader([]byte(b))
			case []batchItem:
				jsonBytes, _ := json.Marshal(b)
				reqBody = bytes.NewReader(jsonBytes)
			default:
				reqBody = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(tt.method, "/api/shorten/batch", reqBody)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, "unexpected status code")

			if tt.wantContent != "" {
				bodyBytes, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(bodyBytes), tt.wantContent, "response should contain shortened URL")
			}
		})
	}
}

func TestPostLongUrl(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		shorten        string
		wantStatusCode int
		wantPrefix     string
	}{
		{
			name:           "#1 - right POST",
			method:         http.MethodPost,
			body:           "https://practicum.yandex.ru/",
			shorten:        "https://lol/",
			wantStatusCode: http.StatusCreated,
			wantPrefix:     "https://lol/",
		},
		{
			name:           "#2 - unknown method",
			method:         http.MethodGet,
			body:           "https://practicum.yandex.ru/",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "#3 — пустое тело",
			method:         http.MethodPost,
			body:           "",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "#4 — валидный урл",
			method:         http.MethodPost,
			body:           "wewq.aa",
			shorten:        "http://localhost:8080/",
			wantStatusCode: http.StatusCreated,
			wantPrefix:     "http://localhost:8080/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewInMemoryStorage()
			router := gin.Default()
			router.HandleMethodNotAllowed = true // <── вот это ключ

			router.POST("/", PostRawURL(store, tt.shorten))

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, "unexpected status code")

			if tt.wantPrefix != "" {
				assert.Truef(t, strings.HasPrefix(string(body), tt.wantPrefix),
					"expected path to start with %q, got %q", tt.wantPrefix, string(body))
			}
		})
	}
}

func TestPostJsonURL(t *testing.T) {
	type args struct {
		method string
		body   map[string]string
	}

	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantContent    string
	}{
		{
			name: "Valid JSON request",
			args: args{
				method: http.MethodPost,
				body:   map[string]string{"url": "https://practicum.yandex.ru/"},
			},
			wantStatusCode: http.StatusCreated,
			wantContent:    `"result":"http://localhost:8080/`, // префикс
		},
		{
			name: "Invalid JSON format",
			args: args{
				method: http.MethodPost,
				body:   nil,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Wrong HTTP method",
			args: args{
				method: http.MethodGet,
				body:   map[string]string{"url": "https://example.com"},
			},
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewInMemoryStorage()
			router := gin.Default()
			router.POST("/api/shorten", PostJSONURL(store, "http://localhost:8080"))

			var reqBody *bytes.Reader
			if tt.args.body != nil {
				jsonBytes, _ := json.Marshal(tt.args.body)
				reqBody = bytes.NewReader(jsonBytes)
			} else {
				reqBody = bytes.NewReader([]byte("invalid_json"))
			}

			req := httptest.NewRequest(tt.args.method, "/api/shorten", reqBody)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, "unexpected status code")

			if tt.wantContent != "" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				assert.Contains(t, buf.String(), tt.wantContent, "response should contain short URL")
			}
		})
	}
}
