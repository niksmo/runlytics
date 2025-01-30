package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HealthCheckHandler struct {
	service HealthCheckService
}

type HealthCheckService interface {
	Check() error
}

func SetHealthCheckHandler(mux *chi.Mux, service HealthCheckService) {
	path := "/ping"
	handler := &HealthCheckHandler{service}
	mux.Get(path, handler.Ping())
	debugLogRegister(path)
}

func (handler *HealthCheckHandler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler.service.Check()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
