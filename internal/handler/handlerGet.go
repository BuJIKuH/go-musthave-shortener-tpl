package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/audit"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

// GetUserURLs возвращает Gin handler, который возвращает список всех URL для текущего пользователя.
//
// Параметры:
//   - s: интерфейс storage.Storage для работы с данными URL
//   - baseURL: базовый адрес коротких ссылок
//
// Логика хендлера:
//  1. Получает userID из контекста (AuthMiddleware).
//  2. Запрашивает все URL пользователя из хранилища.
//  3. Возвращает JSON-массив с полями short_url и original_url.
//
// HTTP ответы:
//   - 200 OK — список URL в формате JSON.
//   - 204 No Content — список пуст.
//   - 401 Unauthorized — отсутствует или пустой userID.
//   - 500 Internal Server Error — ошибка при работе с хранилищем.
func GetUserURLs(s storage.Storage, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("userID")
		if !ok {
			c.String(http.StatusUnauthorized, "no user id")
			return
		}
		userID := v.(string)
		if userID == "" {
			c.String(http.StatusUnauthorized, "invalid token")
			return
		}
		urls, err := s.GetUserURLs(c.Request.Context(), userID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if len(urls) == 0 {
			c.Status(http.StatusNoContent)
			return
		}

		type RespItem struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}

		resp := make([]RespItem, 0, len(urls))
		for _, v := range urls {
			resp = append(resp, RespItem{
				ShortURL:    fmt.Sprintf("%s/%s", baseURL, v.ShortID),
				OriginalURL: v.OriginalURL,
			})
		}
		c.JSON(http.StatusOK, resp)
	}
}

// GetIDURL возвращает Gin handler для редиректа по короткой ссылке.
//
// Параметры:
//   - s: интерфейс storage.Storage для поиска URL по ID
//   - auditSvc: сервис audit.Service для логирования действий пользователей
//
// Логика хендлера:
//  1. Получает параметр "id" из URL.
//  2. Ищет запись в хранилище по ID.
//  3. Если URL найден и не удалён — выполняет редирект на originalURL.
//  4. Отправляет событие в сервис audit для регистрации перехода.
//
// HTTP ответы:
//   - 307 Temporary Redirect — успешный редирект.
//   - 404 Not Found — ID не найден.
//   - 410 Gone — URL помечен как удалён.
func GetIDURL(s storage.Storage, auditSvc *audit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		rec, ok := s.Get(id)
		if !ok || rec == nil {
			c.String(http.StatusNotFound, "id not found")
			return
		}

		if rec.Deleted {
			c.Status(http.StatusGone)
			return
		}

		c.Header("Location", rec.OriginalURL)
		c.Redirect(http.StatusTemporaryRedirect, rec.OriginalURL)

		auditSvc.Notify(
			c.Request.Context(),
			audit.Event{
				TS:     time.Now().Unix(),
				Action: "follow",
				UserID: getUserID(c),
				URL:    rec.OriginalURL,
			})
	}
}
