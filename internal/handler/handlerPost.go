package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func PostBatchURL(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		var req []BatchRequestItem
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		if len(req) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty body"})
			return
		}

		batch := make([]storage.BatchItem, 0, len(req))
		resp := make([]BatchResponseItem, 0, len(req))

		for _, item := range req {
			if strings.TrimSpace(item.OriginalURL) == "" {
				continue
			}

			id, err := shortener.GenerateID()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate short id"})
				return
			}

			batch = append(batch, storage.BatchItem{
				ShortID:     id,
				OriginalURL: item.OriginalURL,
			})

			shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), id)
			resp = append(resp, BatchResponseItem{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})
		}

		if _, _, err := s.SaveBatch(ctx, batch); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save batch"})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func PostJSONURL(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate id"})
			return
		}

		shortID, err := s.Save(ctx, id, req.URL)

		if errors.Is(err, storage.ErrURLExists) {
			shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), shortID)
			c.JSON(http.StatusConflict, ResponseJSON{Result: shortURL})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), shortID)
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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		id, err := shortener.GenerateID()
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to generate id")
			return
		}

		shortID, err := s.Save(ctx, id, originalURL)

		if errors.Is(err, storage.ErrURLExists) {
			shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), shortID)
			c.String(http.StatusConflict, shortURL)
			return
		}
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), shortID)
		c.String(http.StatusCreated, shortURL)
	}
}
