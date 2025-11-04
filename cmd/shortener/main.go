package main

import (
	"log"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/middleware"
	storage2 "github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			config.InitConfig,
			newStorage,
			newRouter,
			NewLogger,
		),
		fx.Invoke(startServer),
	).Run()
}

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return nil, err
	}
	return logger, nil
}

func newStorage(cfg *config.Config, logger *zap.Logger) (storage2.Storage, error) {
	if cfg.FileStoragePath != "" {
		logger.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
		return storage2.NewFileStorage(cfg.FileStoragePath, logger)
	}
	logger.Info("Using in-memory storage")
	return storage2.NewInMemoryStorage(), nil
}

func newRouter(cfg *config.Config, store storage2.Storage, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.Logger(logger),
		middleware.GzipMiddleware(logger),
	)

	r.POST("/", handler.PostRawURL(store, cfg.ShortenAddress))
	r.GET("/:id", handler.GetIDURL(store))
	r.POST("/api/shorten", handler.PostJSONURL(store, cfg.ShortenAddress))
	return r
}

func startServer(cfg *config.Config, r *gin.Engine, logger *zap.Logger) {
	logger.Info("starting server",
		zap.String("address", cfg.Address),
		zap.String("short address", cfg.ShortenAddress))

	if err := r.Run(cfg.Address); err != nil {
		logger.Fatal("Server startup failed", zap.Error(err))
	}
}
