package handler

import (
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

func GetIDURL(s storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		originalURL, ok := s.Get(id)
		if !ok {
			c.String(http.StatusNotFound, "id not found")
			return
		}
		c.Header("Location", originalURL)
		c.Redirect(http.StatusTemporaryRedirect, originalURL)
	}
}
