package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func PostLongURL(storage map[string]string, shortURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
		if c.GetHeader("Content-Type") != "text/plain" {
			c.String(http.StatusBadRequest, "invalid content type")
			return
		}

		body, err := c.GetRawData()
		if err != nil || len(body) == 0 {
			c.String(http.StatusBadRequest, "empty body")
			return
		}

		originalURL := strings.TrimSpace(string(body))
		u, err := url.ParseRequestURI(originalURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			c.String(http.StatusBadRequest, "invalid url")
			return
		}

		const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 8)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		id := string(b)
		storage[id] = originalURL
		finishURL := fmt.Sprintf("%s/%s", shortURL, id)

		// Добавляем заголовки
		c.Header("Content-Type", "text/plain")
		c.Header("Content-Length", fmt.Sprint(len(finishURL)))

		// Отправляем ответ
		c.String(http.StatusCreated, finishURL)
	}
}
