package main

import (
	"log"
	"net/http"

	"github.com/niksmo/runlytics/internal/config"
	"github.com/niksmo/runlytics/internal/handler"
)

type ServerConfig interface {
	Addr() string
}

func main() {
	log.Println("Bootstrap server...")
	config := config.New()
	mux := http.NewServeMux()

	handler.NewUpdate(mux)

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
	log.Println("Listen host", s.Addr)
	return s.ListenAndServe()
}
