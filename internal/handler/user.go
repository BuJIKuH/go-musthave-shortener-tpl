// Package handler содержит HTTP-хендлеры и вспомогательные функции для работы с запросами.
package handler

import "github.com/gin-gonic/gin"

// getUserID извлекает userID из контекста Gin.
//
// Возвращает пустую строку, если userID отсутствует или имеет некорректный тип.
func getUserID(c *gin.Context) string {
	v, ok := c.Get("userID")
	if !ok {
		return ""
	}

	id, ok := v.(string)
	if !ok {
		return ""
	}

	return id
}
