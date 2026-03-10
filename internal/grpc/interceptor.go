// Package grpcserver содержит gRPC сервер и middleware.
package grpcserver

import (
	"context"
	"errors"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var ErrUnauthenticated = errors.New("unauthenticated")

// AuthUnaryInterceptor создает unary interceptor для проверки токена Authorization.
// Все gRPC вызовы будут проверять заголовок "authorization".
// Если токен валиден, userID сохраняется в контекст для дальнейшего использования.
func AuthUnaryInterceptor(am *auth.Manager, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		var token string
		if t, exists := md["authorization"]; exists && len(t) > 0 {
			token = t[0]
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		} else {
			return nil, status.Error(codes.Unauthenticated, "authorization header missing")
		}

		userID, err := am.ParseToken(token, logger)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// 🔹 Используем правильный ключ для контекста
		ctx = context.WithValue(ctx, ctxKeyUserID{}, userID)
		return handler(ctx, req)
	}
}

// ctxKeyUserID используется для хранения userID в контексте
type ctxKeyUserID struct{}

// UserIDFromContext извлекает userID из контекста
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyUserID{}).(string)
	return id, ok
}
