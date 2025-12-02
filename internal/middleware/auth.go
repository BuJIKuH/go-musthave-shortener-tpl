package middleware

import (
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const userIDKey = "userID"

func generateToken(am *auth.Manager, logger *zap.Logger) (newID, newToken string, err error) {
	newID = uuid.NewString()
	token, err := am.GenerateToken(newID)
	if err != nil {
		logger.Error("failed to generate token", zap.Error(err))
		return
	}
	return newID, token, nil
}

func AuthMiddleware(am *auth.Manager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie, err := c.Cookie("auth_token")

		if err != nil || cookie == "" {
			newID, token, err := generateToken(am, logger)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			c.Set(userIDKey, newID)
			c.Next()
			return
		}

		userID, err := am.ParseToken(cookie, logger)
		if err != nil || userID == "" {
			newID, token, err := generateToken(am, logger)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			c.Set(userIDKey, newID)
			c.Next()
			return
		}
		logger.Info("user id", zap.String("user_id", userID), zap.String("path", c.Request.URL.Path))

		c.Set(userIDKey, userID)
		c.Next()
	}
}
