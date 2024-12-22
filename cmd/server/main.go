package main

import (
	"log"
	"net/http"

	"github.com/niksmo/runlytics/internal/config"
	"github.com/niksmo/runlytics/internal/repository"
	"github.com/niksmo/runlytics/internal/server"
)

type ServerConfig interface {
	Addr() string
}

func main() {
	log.Println("Bootstrap server")

	config := config.NewServerConfig()
	mux := http.NewServeMux()
	storage := repository.NewMemStorage()
	server.NewHandler(mux, storage)

	err := run(config, mux)
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
