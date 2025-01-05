package router

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
)

func SetValueRoute(r *chi.Mux, repo server.RepositoryRead) {
	h := &valueHandler{repo}
	r.Route("/value", func(r chi.Router) {
		r.Get("/", h.getHandleFunc())
		log.Println("Register endpoint: /value/{type}/{name}")
	})
}

type valueHandler struct {
	repo server.RepositoryRead
}

func (h *valueHandler) getHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
