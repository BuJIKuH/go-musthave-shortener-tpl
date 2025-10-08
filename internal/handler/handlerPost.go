package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service/shortener"
	"github.com/gin-gonic/gin"
)

type Storage interface {
	Save(id, url string)
	Get(id string) (string, bool)
}

func PostLongURL(s Storage, shortURL string) gin.HandlerFunc {
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

		id, err := shortener.GenerateID()
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to generate id")
		}

		s.Save(id, originalURL)

		finishURL := fmt.Sprintf("%s/%s", strings.TrimRight(shortURL, "/"), id)
		if !strings.HasPrefix(shortURL, "http://") && !strings.HasPrefix(finishURL, "https://") {
			finishURL = "http://" + finishURL
		}

		// Добавляем заголовки
		c.Header("Content-Type", "text/plain")
		c.Header("Content-Length", fmt.Sprint(len(finishURL)))

		// Отправляем ответ
		c.String(http.StatusCreated, finishURL)
	}
}
