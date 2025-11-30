package handler

import (
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

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
