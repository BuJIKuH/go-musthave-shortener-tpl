package main

import (
	"context"
	"log"
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

func NewAuditService(cfg *config.Config) *audit.Service {
	var observers []audit.Observer

	if cfg.AuditFile != "" {
		fo, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			log.Fatalf("failed to init audit file observer: %v", err)
		}
		observers = append(observers, fo)
	}

	if cfg.AuditURL != "" {
		observers = append(observers, audit.NewHTTPObserver(cfg.AuditURL))
	}

	return audit.NewService(observers...)
}

func NewDeleter(lc fx.Lifecycle, store storage.Storage, logger *zap.Logger) *service.Deleter {
	d := service.NewDeleter(store.MarkDeleted)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down deleter gracefully...")
			d.Close()
			return nil
		},
	})

	return d
}

func NewAuthManager(cfg *config.Config) *auth.Manager {
	return auth.NewManager(cfg.AuthSecret)
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

func newRouter(
	cfg *config.Config,
	store storage.Storage,
	am *auth.Manager,
	deleter *service.Deleter,
	auditSvc *audit.Service,
	logger *zap.Logger) *gin.Engine {
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
					logger.Fatal("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server...")
			return srv.Shutdown(ctx)
		},
	})
}
