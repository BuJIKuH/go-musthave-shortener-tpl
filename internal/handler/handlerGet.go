package handler

import (
	"net/http"
	"strings"
)

func GetIdUrl(storage map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if r.Method != http.MethodGet || path == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		origninalUrl, ok := storage[path]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Location", origninalUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

}
