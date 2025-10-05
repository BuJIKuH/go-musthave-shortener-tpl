package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIdUrl(t *testing.T) {
	type args struct {
		storage map[string]string
		method  string
		path    string
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantUrl        string
	}{
		{
			name: "#1 - GET",
			args: args{
				storage: map[string]string{
					"sdasda": "https://practicum.yandex.ru/",
				},
				method: http.MethodGet,
				path:   "/sdasda",
			},
			wantStatusCode: http.StatusTemporaryRedirect,
			wantUrl:        "https://practicum.yandex.ru/",
		},
		{
			name: "#2 - unknown method",
			args: args{
				storage: map[string]string{
					"sdasda": "https://practicum.yandex.ru/",
				},
				method: http.MethodPost,
				path:   "/sdasda",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "#3 — пустое тело",
			args: args{
				storage: map[string]string{
					"sdasda": "https://practicum.yandex.ru/",
				},
				method: http.MethodGet,
				path:   "/",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "#4 — невалидный path",
			args: args{
				storage: map[string]string{
					"sdasda": "https://practicum.yandex.ru/",
				},
				method: http.MethodGet,
				path:   "/saasda",
			},
			wantStatusCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.args.method, tt.args.path, strings.NewReader(tt.args.path))
			w := httptest.NewRecorder()

			GetIdUrl(tt.args.storage)(w, req)
			resp := w.Result()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if tt.wantUrl != "" {
				assert.Equal(t, tt.wantUrl, resp.Header.Get("Location"))
			}

			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Logf("Response path: %s", strings.TrimSpace(string(body)))

		})
	}
}
