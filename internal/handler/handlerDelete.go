package handler

import (
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

// DeleteUserURLs возвращает Gin handler для асинхронного удаления URL пользователя.
//
// Этот хендлер:
//  1. Извлекает userID из контекста (устанавливается AuthMiddleware).
//  2. Парсит JSON-массив идентификаторов URL для удаления.
//  3. Отправляет задачу на удаление в сервис Deleter.
//  4. Возвращает статус 202 Accepted.
//
// Ответы:
//   - 202 Accepted — задача принята в обработку.
//   - 400 Bad Request — неверный формат JSON.
//   - 401 Unauthorized — userID отсутствует в контексте.
func DeleteUserURLs(s storage.Storage, d *service.Deleter) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("userID")
		if !ok {
			c.Status(http.StatusUnauthorized)
			return
		}
		userID := v.(string)
		if userID == "" {
			c.Status(http.StatusUnauthorized)
			return
		}

		var ids []string
		if err := c.BindJSON(&ids); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		d.Enqueue(service.DeleteTask{
			UserID: userID,
			IDs:    ids,
		})

		c.Status(http.StatusAccepted)
	}
}
