package handler

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func PostLongUrl(storage map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.Method != "POST" {
			log.Println(w, "bad request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		defer r.Body.Close()

		originalUrl := string(body)

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
