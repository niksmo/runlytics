package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/storage"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 8080
)

func main() {
	log.Println("Bootstrap server")

	router := http.NewServeMux()
	storage := storage.NewMemStorage()
	server.NewHandler(router, storage)

	err := run(router)
	if err != nil {
		log.Fatal(err)
	}

}

func run(handler *http.ServeMux) error {
	addr := defaultHost + ":" + strconv.Itoa(defaultPort)
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Println("Listen", s.Addr)
	return s.ListenAndServe()
}
