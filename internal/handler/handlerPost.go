package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config/storage"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service/shortener"
	"github.com/gin-gonic/gin"
)

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

func PostJsonURL(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": http.StatusText(http.StatusMethodNotAllowed)})
			return
		}

		var req RequestJSON
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		id, err := shortener.GenerateID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		s.Save(id, req.URL)

		shortURL := baseURL + "/" + id

		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusCreated, ResponseJSON{shortURL})
	}
}

func PostRawURL(s storage.Storage, shortURL string) gin.HandlerFunc {
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
			return
		}

		s.Save(id, originalURL)

		finishURL := fmt.Sprintf("%s/%s", strings.TrimRight(shortURL, "/"), id)
		if !strings.HasPrefix(shortURL, "http://") && !strings.HasPrefix(finishURL, "https://") {
			finishURL = "http://" + finishURL
		}

		c.Header("Content-Type", "text/plain")
		c.Header("Content-Length", fmt.Sprint(len(finishURL)))

		c.String(http.StatusCreated, finishURL)
	}
}
