// Package main запускает сервис сокращения URL с поддержкой аудита,
// логирования, pprof и встроенной авторизации.
package main

import (
	"context"
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/audit"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/auth"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/middleware"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

func main() {
	fx.New(
		fx.Provide(
			config.InitConfig,
			newStorage,
			newRouter,
			NewLogger,
			NewAuthManager,
			NewDeleter,
			NewAuditService,
		),
		fx.Invoke(startServer),
	).Run()
}

// NewAuditService создает сервис аудита с указанными наблюдателями.
func NewAuditService(cfg *config.Config, logger *zap.Logger) *audit.Service {
	observers := make([]audit.Observer, 0)

	if cfg.AuditFile != "" {
		fo, err := audit.NewFileObserver(cfg.AuditFile, logger)
		if err != nil {
			logger.Error("failed to init audit file observer", zap.Error(err))
		} else {
			observers = append(observers, fo)
		}
	}

	if cfg.AuditURL != "" {
		observers = append(observers, audit.NewHTTPObserver(cfg.AuditURL, logger))
	}

	return audit.NewService(logger, observers...)
}

// NewDeleter создает сервис Deleter для пометки URL как удаленных.
func NewDeleter(lc fx.Lifecycle, store storage.Storage, logger *zap.Logger) *service.Deleter {
	d := service.NewDeleter(store.MarkDeleted)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down deleter gracefully")
			d.Close()
			return nil
		},
	})

	return d
}

// NewAuthManager создает менеджер авторизации.
func NewAuthManager(cfg *config.Config) *auth.Manager {
	return auth.NewManager(cfg.AuthSecret)
}

// NewLogger инициализирует Zap Logger в production-режиме.
func NewLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

// newStorage создает и возвращает хранилище для URL.
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

// newRouter создает и настраивает маршрутизатор Gin.
func newRouter(
	cfg *config.Config,
	store storage.Storage,
	am *auth.Manager,
	deleter *service.Deleter,
	auditSvc *audit.Service,
	logger *zap.Logger,
) *gin.Engine {

	r := gin.New()
	r.Use(
		middleware.Logger(logger),
		middleware.GzipMiddleware(logger),
		middleware.AuthMiddleware(am, logger),
	)

	r.POST("/", handler.PostRawURL(store, cfg.ShortenAddress, auditSvc))
	r.GET("/:id", handler.GetIDURL(store, auditSvc))
	r.POST("/api/shorten", handler.PostJSONURL(store, cfg.ShortenAddress, auditSvc))
	r.GET("/ping", handler.PingHandler(store))
	r.POST("/api/shorten/batch", handler.PostBatchURL(store, cfg.ShortenAddress))
	r.GET("/api/user/urls", handler.GetUserURLs(store, cfg.ShortenAddress))
	r.DELETE("/api/user/urls", handler.DeleteUserURLs(store, deleter))

	return r
}

// startServer запускает HTTP сервер и pprof сервер.
func startServer(lc fx.Lifecycle, cfg *config.Config, r *gin.Engine, logger *zap.Logger) {
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting HTTP server", zap.String("address", cfg.Address))

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server error", zap.Error(err))
				}
			}()

			go func() {
				logger.Info("Starting pprof server on :6060")
				if err := http.ListenAndServe("localhost:6060", nil); err != nil {
					logger.Error("pprof server error", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
