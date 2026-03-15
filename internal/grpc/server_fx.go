// Package grpcserver содержит запуск gRPC сервера через fx lifecycle.
package grpcserver

import (
	"context"
	"net"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/auth"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	shortenerpb "github.com/BuJIKuH/go-musthave-shortener-tpl/internal/proto/shortener"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // 🔹 импорт для reflection
)

// Params описывает зависимости для запуска gRPC сервера.
type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Store     storage.Storage
	Auth      *auth.Manager
	Logger    *zap.Logger
	Config    *config.Config
}

// RunGRPCServer регистрирует и запускает gRPC сервер.
func RunGRPCServer(p Params) {
	// создаем gRPC сервер с interceptor для авторизации
	server := grpc.NewServer(
		grpc.UnaryInterceptor(AuthUnaryInterceptor(p.Auth, p.Logger)),
	)

	grpcService := NewGRPCServer(
		p.Store,
		p.Logger,
	)

	shortenerpb.RegisterShortenerServiceServer(server, grpcService)

	// включаем reflection, чтобы grpcurl и Insomnia могли видеть сервисы
	reflection.Register(server)

	addr := p.Config.GRPCAddress

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		p.Logger.Fatal("failed to listen gRPC", zap.Error(err))
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				p.Logger.Info("Starting gRPC server", zap.String("address", addr))
				if err := server.Serve(lis); err != nil {
					p.Logger.Error("gRPC server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping gRPC server")
			server.GracefulStop()
			return nil
		},
	})
}
