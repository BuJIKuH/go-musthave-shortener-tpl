package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
)

var storage = make(map[string]string)

func main() {
	r := gin.Default()

	r.POST("/", handler.PostLongUrl(storage))
	r.GET("/:id", handler.GetIdUrl(storage))

	log.Println("server is running on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Panic(err)
	}
}
