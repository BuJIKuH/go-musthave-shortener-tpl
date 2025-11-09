package main

import (
	"log"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/middleware"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"

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

func newStorage(cfg *config.Config, logger *zap.Logger) (storage.Storage, error) {
	if cfg.DatabaseDSN != "" {
		if err := storage.RunMigrations(cfg.DatabaseDSN, logger); err != nil {
			logger.Error("can't initialize database migrations", zap.Error(err))
		}

		dbStore, err := storage.NewDBStorage(cfg.DatabaseDSN, logger)
		if err == nil {
			logger.Info("Using PostgreSQL storage")
			return dbStore, nil
		}
		logger.Error("Failed to connect to PostgreSQL, falling back to file storage", zap.Error(err))
	}

	if cfg.FileStoragePath != "" {
		logger.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
		return storage.NewFileStorage(cfg.FileStoragePath, logger)
	}
	logger.Info("Using in-memory storage")
	return storage.NewInMemoryStorage(), nil
}

func newRouter(cfg *config.Config, store storage.Storage, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.Logger(logger),
		middleware.GzipMiddleware(logger),
	)

	r.POST("/", handler.PostRawURL(store, cfg.ShortenAddress))
	r.GET("/:id", handler.GetIDURL(store))
	r.POST("/api/shorten", handler.PostJSONURL(store, cfg.ShortenAddress))
	r.GET("/ping", handler.PingHandler(store))
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
