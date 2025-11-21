package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Manager struct {
	Secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{[]byte(secret)}
}

func (m *Manager) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

func (m *Manager) ParseToken(tokenStr string, logger *zap.Logger) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			logger.Error("unexpected signing method", zap.String("Method", "HS256"))
			return nil, errors.New("unexpected signing method")
		}
		return m.Secret, nil
	})
	if err != nil || !token.Valid {
		logger.Error("invalid token", zap.Error(err))
		return "", errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	if sub == "" {
		logger.Error("failed to extract sub from token, no userID", zap.String("sub", sub))
		return "", errors.New("invalid sub claim")
	}

	return sub, nil
}
