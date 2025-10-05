package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostLongUrl(t *testing.T) {
	type args struct {
		storage map[string]string
		method  string
		body    string
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
				storage: make(map[string]string),
				method:  http.MethodPost,
				body:    "https://practicum.yandex.ru/",
			},
			wantStatusCode: http.StatusCreated,
			wantPrefix:     "http://localhost:8080/",
		},
		{
			name: "#2 - unknown method",
			args: args{
				storage: make(map[string]string),
				method:  http.MethodGet,
				body:    "https://practicum.yandex.ru/",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "#3 — пустое тело",
			args: args{
				storage: make(map[string]string),
				method:  http.MethodPost,
				body:    "",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "#4 — невалидный урл",
			args: args{
				storage: make(map[string]string),
				method:  http.MethodPost,
				body:    "ww.wewq.aa",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "#5 — невалидный урл",
			args: args{
				storage: make(map[string]string),
				method:  http.MethodPost,
				body:    "www.practicum.yandex.ru",
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.args.method, "/", strings.NewReader(tt.args.body))
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			handler := PostLongUrl(tt.args.storage)
			handler(w, req)

			resp := w.Result()
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
