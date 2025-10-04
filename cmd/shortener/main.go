package main

import (
	"log"
	"net/http"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/handler"
)

var storage = make(map[string]string)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.PostLongUrl(storage))
	mux.HandleFunc("/{id}", handler.GetIdUrl(storage))
	if err := run(mux); err != nil {
		log.Panic(err)
	}
}

func run(mux http.Handler) error {
	log.Println("server is running on port 8080")
	return http.ListenAndServe("localhost:8080", mux)
}
