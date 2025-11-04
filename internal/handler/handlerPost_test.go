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

func TestPostLongUrl(t *testing.T) {
	type args struct {
		storage *storage.InMemoryStorage
		method  string
		body    string
		shorten string
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantPrefix     string
	}{
		{
			name: "#1 - right POST",
			args: args{
				storage: storage.NewInMemoryStorage(),
				method:  http.MethodPost,
				body:    "https://practicum.yandex.ru/",
				shorten: "https://lol/",
			},
			wantStatusCode: http.StatusCreated,
			wantPrefix:     "https://lol/",
		},
		{
			name: "#2 - unknown method",
			args: args{
				storage: storage.NewInMemoryStorage(),
				method:  http.MethodGet,
				body:    "https://practicum.yandex.ru/",
			},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name: "#3 — пустое тело",
			args: args{
				storage: storage.NewInMemoryStorage(),
				method:  http.MethodPost,
				body:    "",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "#4 — валидный урл",
			args: args{
				storage: storage.NewInMemoryStorage(),
				method:  http.MethodPost,
				body:    "wewq.aa",
				shorten: "http://localhost:8080/",
			},
			wantStatusCode: http.StatusCreated,
			wantPrefix:     "http://localhost:8080/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(tt.args.method, "/", strings.NewReader(tt.args.body))
			req.Header.Set("Content-Type", "text/plain")

			c.Request = req

			handler := PostRawURL(tt.args.storage, tt.args.shorten)
			handler(c)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			gotBody := string(body)

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, "unexpected status code")

			if tt.wantPrefix != "" {
				assert.Truef(t, strings.HasPrefix(gotBody, tt.wantPrefix),
					"expected path to start with %q, got %q", tt.wantPrefix, gotBody)
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
