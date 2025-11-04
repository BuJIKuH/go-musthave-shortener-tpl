package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service/shortener"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

func PostJSONURL(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RequestJSON
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		if strings.TrimSpace(req.URL) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		id, err := shortener.GenerateID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate short id"})
			return
		}

		s.Save(id, req.URL)

		shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), id)
		c.JSON(http.StatusCreated, ResponseJSON{Result: shortURL})
	}
}

func PostRawURL(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), id)
		c.String(http.StatusCreated, shortURL)
	}
}
