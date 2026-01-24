package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/audit"
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

// PostBatchURL возвращает Gin handler для массового сокращения URL.
//
// Параметры:
//   - s: интерфейс storage.Storage для сохранения URL
//   - baseURL: базовый адрес для формирования коротких ссылок
//
// Логика хендлера:
//  1. Проверяет наличие userID в контексте.
//  2. Декодирует JSON-массив BatchRequestItem.
//  3. Генерирует короткие ID для каждого URL.
//  4. Сохраняет batch в хранилище.
//  5. Возвращает JSON-массив BatchResponseItem с короткими ссылками.
//
// HTTP ответы:
//   - 201 Created — успешно сохранён batch.
//   - 400 Bad Request — пустой массив или некорректный JSON.
//   - 401 Unauthorized — отсутствует userID.
//   - 500 Internal Server Error — ошибка генерации ID или сохранения batch.
func PostBatchURL(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		u, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userID"})
			return
		}
		userID := u.(string)

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

		if _, _, err := s.SaveBatch(ctx, userID, batch); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save batch"})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// PostJSONURL возвращает Gin handler для сокращения одного URL через JSON.
//
// Параметры:
//   - s: интерфейс storage.Storage
//   - baseURL: базовый адрес коротких ссылок
//   - auditSvc: сервис audit.Service для логирования действий
//
// Логика хендлера:
//  1. Декодирует JSON с полем "url".
//  2. Генерирует короткий ID.
//  3. Сохраняет URL в хранилище.
//  4. Возвращает JSON с полем "result" — короткая ссылка.
//  5. Отправляет событие в audit сервис.
//
// HTTP ответы:
//   - 201 Created — успешное создание новой короткой ссылки.
//   - 409 Conflict — URL уже существует, возвращается существующая короткая ссылка.
//   - 400 Bad Request — пустой или некорректный JSON.
//   - 500 Internal Server Error — ошибка генерации ID или сохранения URL.
func PostJSONURL(s storage.Storage, baseURL string, auditSvc *audit.Service) gin.HandlerFunc {
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
		originalURL := strings.TrimSpace(req.URL)

		u, _ := c.Get("userID")
		userID := u.(string)

		id, err := shortener.GenerateID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate id"})
			return
		}

		shortID, err := s.Save(ctx, userID, id, req.URL)

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

		auditSvc.Notify(
			c.Request.Context(),
			audit.Event{TS: time.Now().Unix(),
				Action: "shorten",
				UserID: getUserID(c),
				URL:    originalURL})
	}
}

// PostRawURL возвращает Gin handler для сокращения одного URL из текста.
//
// Параметры:
//   - s: интерфейс storage.Storage
//   - baseURL: базовый адрес коротких ссылок
//   - auditSvc: сервис audit.Service для логирования действий
//
// Логика хендлера:
//  1. Проверяет Content-Type "text/plain".
//  2. Читает тело запроса и проверяет на пустоту.
//  3. Генерирует короткий ID.
//  4. Сохраняет URL в хранилище.
//  5. Возвращает короткую ссылку как plain text.
//  6. Отправляет событие в audit сервис.
//
// HTTP ответы:
//   - 201 Created — успешно создана короткая ссылка.
//   - 409 Conflict — URL уже существует, возвращается существующая короткая ссылка.
//   - 400 Bad Request — пустое тело или некорректный Content-Type.
//   - 500 Internal Server Error — ошибка генерации ID или сохранения URL.
func PostRawURL(s storage.Storage, baseURL string, auditSvc *audit.Service) gin.HandlerFunc {
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

		u, _ := c.Get("userID")
		userID := u.(string)

		id, err := shortener.GenerateID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate id"})
			return
		}

		shortID, err := s.Save(ctx, userID, id, originalURL)

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

		auditSvc.Notify(
			c.Request.Context(),
			audit.Event{
				TS:     time.Now().Unix(),
				Action: "shorten",
				UserID: getUserID(c),
				URL:    originalURL})
	}
}
