package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/router"
	"github.com/niksmo/runlytics/internal/storage"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 8080
)

func main() {
	log.Println("Bootstrap server")
	mux := chi.NewRouter()

	storage := storage.NewMemStorage()
	router.SetUpdateRoute(mux, storage)
	router.SetValueRoute(mux, storage)

	log.Fatal(run(mux))

}

func run(handler *chi.Mux) error {
	addr := defaultHost + ":" + strconv.Itoa(defaultPort)
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Println("Listen", s.Addr)
	return s.ListenAndServe()
}
