package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/pkg/di"
)

// HealthCheckHandler works with service and provides Ping method.
type HealthCheckHandler struct {
	service di.HealthCheckService
}

// SetHealthCheckHandler sets Ping handler to "/ping" path.
func SetHealthCheckHandler(mux *chi.Mux, service di.HealthCheckService) {
	path := "/ping"
	handler := &HealthCheckHandler{service}
	mux.Get(path, handler.Ping())
	debugLogRegister(path)
}

// Ping check internal server components.
//
// Possible responses:
//
//   - 200 all components is up
//   - 500 internal error
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
