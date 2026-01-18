package handler

import "github.com/gin-gonic/gin"

func getUserID(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
