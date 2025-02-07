package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type BatchUpdateHandler struct {
	service   di.BatchUpdateService
	validator di.SchemeVerifier
}

func SetBatchUpdateHandler(
	mux *chi.Mux,
	service di.BatchUpdateService,
	validator di.SchemeVerifier,
) {
	path := "/updates"
	handler := &BatchUpdateHandler{service, validator}
	mux.Route(path, func(r chi.Router) {
		batchUpdate := "/"
		r.With(middleware.AllowJSON).Post(path+batchUpdate, handler.batchUpdate())
		debugLogRegister(path + batchUpdate)
	})
}

func (handler *BatchUpdateHandler) batchUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var scheme metrics.MetricsBatchUpdate
		if err := decodeJSON(r, &scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := handler.validator.VerifyScheme(&scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := handler.service.BatchUpdate(r.Context(), scheme)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
