package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ValueHandler struct {
	service ValueService
}

type ValueService interface {
}

func SetReadHandler(mux *chi.Mux, service ValueService) {
	handler := &ValueHandler{service}
	mux.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handler.get())
		debugLogRegister("/value/{type}/{name}")
	})
}

func (handler *ValueHandler) get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
