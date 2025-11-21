package handler

import (
	"fmt"
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
)

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
