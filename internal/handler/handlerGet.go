package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetIdUrl(storage map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		origninalUrl, ok := storage[id]
		if !ok {
			c.String(http.StatusNotFound, "id not found")
			return
		}
		c.Header("Location", origninalUrl)
		c.Redirect(http.StatusTemporaryRedirect, origninalUrl)
	}
}
