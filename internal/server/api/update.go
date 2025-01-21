package api

import (
	"bytes"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UpdateHandler struct {
	service UpdateService
}

type UpdateService interface {
	Update()
}

func SetUpdateHandler(mux *chi.Mux, service UpdateService) {
	handler := &UpdateHandler{service}
	mux.Route("/update", func(r chi.Router) {
		r.Post("/", handler.post())
		debugLogRegister("/update/")
	})
}

func (handler *UpdateHandler) post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		buf := bytes.NewBufferString("Hello world!")
		buf.WriteTo(w)
	}
}
