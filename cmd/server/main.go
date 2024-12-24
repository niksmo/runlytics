package main

import (
	"log"
	"net/http"

	"github.com/niksmo/runlytics/internal/config"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/storage"
)

type ServerConfig interface {
	Addr() string
}

func main() {
	log.Println("Bootstrap server")

	config := config.NewServerConfig()
	router := http.NewServeMux()
	storage := storage.NewMemStorage()
	server.NewHandler(router, storage)

	err := run(config, router)
	if err != nil {
		log.Fatal(err)
	}

}

func run(c ServerConfig, handler *http.ServeMux) error {
	s := http.Server{
		Addr:    c.Addr(),
		Handler: handler,
	}
	log.Println("Listen", s.Addr)
	return s.ListenAndServe()
}
