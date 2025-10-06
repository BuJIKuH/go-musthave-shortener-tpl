package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
		wantURL        string
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
			wantURL:        "https://practicum.yandex.ru/",
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
			wantStatusCode: http.StatusNotFound,
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
			wantStatusCode: http.StatusNotFound,
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
			router := gin.New()
			router.GET("/:id", GetIDURL(tt.args.storage))

			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			//GetIDURL(tt.args.storage)(c)
			//resp := w.Result()

			assert.Equal(t, tt.wantStatusCode, w.Code, "unexpected status code")

			if tt.wantURL != "" {
				assert.Equal(t, tt.wantURL, w.Header().Get("Location"))
			}

		})
	}
}
