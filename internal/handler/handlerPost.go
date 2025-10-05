package handler

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func PostLongUrl(storage map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != http.MethodPost || r.Body == nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || len(body) == 0 {
			http.Error(w, "empty path", http.StatusBadRequest)
			return
		}

		originalUrl := strings.TrimSpace(string(body))

		u, err := url.ParseRequestURI(originalUrl)
		log.Println(u)
		if err != nil || u.Scheme == "" || u.Host == "" {
			http.Error(w, "Invalid url", http.StatusBadRequest)
			return
		}

		const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		rand.Seed(time.Now().UnixNano())
		b := make([]byte, 8)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		storage[string(b)] = originalUrl

		shortUrl := fmt.Sprintf("http://localhost:8080/%s", string(b))
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", fmt.Sprint(len(shortUrl)))
		w.WriteHeader(http.StatusCreated)

		fmt.Fprint(w, shortUrl)

		log.Println(originalUrl, string(b))
	}

}
