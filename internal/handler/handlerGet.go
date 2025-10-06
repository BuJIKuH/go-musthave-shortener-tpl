package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetIDURL(storage map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		originalURL, ok := storage[id]
		if !ok {
			c.String(http.StatusNotFound, "id not found")
			return
		}
		c.Header("Location", originalURL)
		c.Redirect(http.StatusTemporaryRedirect, originalURL)
	}
}
