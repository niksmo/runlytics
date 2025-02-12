package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/pkg/di"
)

type HealthCheckHandler struct {
	service di.HealthCheckService
}

func SetHealthCheckHandler(mux *chi.Mux, service di.HealthCheckService) {
	path := "/ping"
	handler := &HealthCheckHandler{service}
	mux.Get(path, handler.Ping())
	debugLogRegister(path)
}

func (handler *HealthCheckHandler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler.service.Check(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
