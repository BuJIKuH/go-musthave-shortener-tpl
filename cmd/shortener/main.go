package main

import (
	"log"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config/db"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.InitConfig,
			db.NewInMemoryStorage,
			newRouter,
		),
		fx.Invoke(startServer),
	).Run()
}

func newRouter(cfg *config.Config, store *db.InMemoryStorage) *gin.Engine {
	r := gin.Default()
	r.POST("/", handler.PostLongURL(store, cfg.ShortenAddress))
	r.GET("/:id", handler.GetIDURL(store))
	return r
}

func startServer(cfg *config.Config, r *gin.Engine) {
	log.Printf("Starting server on %s with base URL %s\n", cfg.Address, cfg.ShortenAddress)
	if err := r.Run(cfg.Address); err != nil {
		log.Panic(err)
	}
}
