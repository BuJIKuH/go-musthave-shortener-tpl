package main

import (
	"log"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config/storage"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			config.InitConfig,
			storage.NewInMemoryStorage,
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

func newRouter(cfg *config.Config, store *storage.InMemoryStorage, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(logger))

	r.POST("/", handler.PostLongURL(store, cfg.ShortenAddress))
	r.GET("/:id", handler.GetIDURL(store))
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
