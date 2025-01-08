package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/router"
	"github.com/niksmo/runlytics/internal/storage"
)

func main() {
	parseFlags()

	log.Println("Bootstrap server")
	mux := chi.NewRouter()

	storage := storage.NewMemStorage()
	router.SetMainRoute(mux, storage)
	router.SetUpdateRoute(mux, storage)
	router.SetValueRoute(mux, storage)

	log.Fatal(run(mux))
}

func run(handler *chi.Mux) error {
	s := http.Server{
		Addr:    fmt.Sprint(flagAddr),
		Handler: handler,
	}
	log.Println("Listen", s.Addr)
	return s.ListenAndServe()
}
