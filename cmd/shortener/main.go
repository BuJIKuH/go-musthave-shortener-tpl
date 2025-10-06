package main

import (
	"log"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config/config"
	"github.com/gin-gonic/gin"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
)

var storage = make(map[string]string)

func main() {
	cfg := config.InitConfig()

	log.Printf("Starting server on %s with base URL %s\n", cfg.Address, cfg.ShortenAddress)

	r := gin.Default()

	r.POST("/", handler.PostLongURL(storage, cfg.ShortenAddress))
	r.GET("/:id", handler.GetIDURL(storage))

	log.Println("server is running on port: ", cfg.ShortenAddress)
	if err := r.Run(cfg.Address); err != nil {
		log.Panic(err)
	}
}
