package handler

import (
	"net"
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

// StatsResponse описывает JSON-ответ эндпоинта статистики сервиса.
type StatsResponse struct {
	// URLs — общее количество сокращённых URL в сервисе.
	URLs int `json:"urls"`
	// Users — количество пользователей, имеющих сохранённые URL.
	Users int `json:"users"`
}

// GetStats возвращает Gin handler для получения внутренней статистики сервиса.
//
// Эндпоинт: GET /api/internal/stats
//
// Handler выполняет следующие шаги:
//  1. Проверяет, что в конфигурации задана доверенная подсеть trusted_subnet.
//  2. Извлекает IP клиента из заголовка X-Real-IP.
//  3. Проверяет, что IP клиента принадлежит доверенной подсети.
//  4. Получает статистику количества URL и пользователей из хранилища.
//  5. Возвращает JSON-ответ со статистикой.
//
// HTTP ответы:
//   - 200 OK — статистика успешно получена.
//   - 403 Forbidden — доступ запрещён (IP не из доверенной подсети или subnet не задан).
func GetStats(s storage.Storage, cfg *config.Config) gin.HandlerFunc {
	var subnet *net.IPNet

	if cfg.TrustedSubnet != "" {
		_, sn, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err == nil {
			subnet = sn
		}
	}

	return func(c *gin.Context) {

		if subnet == nil {
			c.Status(http.StatusForbidden)
			return
		}

		ipStr := c.GetHeader("X-Real-IP")
		if ipStr == "" {
			c.Status(http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			c.Status(http.StatusForbidden)
			return
		}

		if !subnet.Contains(ip) {
			c.Status(http.StatusForbidden)
			return
		}

		resp := StatsResponse{
			URLs:  s.URLsCount(),
			Users: s.UsersCount(),
		}

		c.JSON(http.StatusOK, resp)
	}
}
