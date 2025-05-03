package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/jsonhttp"
	"github.com/niksmo/runlytics/pkg/metrics"
)

// BatchUpdateHandler working with service and provides BatchUpdate method.
type BatchUpdateHandler struct {
	service di.BatchUpdateService
}

// SetBatchUpdateHandler sets BatchUpdate handler to "/updates" path.
//
// Allows only JSON media type.
func SetBatchUpdateHandler(mux *chi.Mux, service di.BatchUpdateService) {
	path := "/updates"
	handler := &BatchUpdateHandler{service}
	mux.Route(path, func(r chi.Router) {
		batchUpdate := "/"
		r.With(middleware.AllowJSON).Post(batchUpdate, handler.BatchUpdate())
		debugLogRegister(path + batchUpdate)
	})
}

// BatchUpdate reads metrics list from request for update.
func (h *BatchUpdateHandler) BatchUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ml metrics.MetricsList
		if err := jsonhttp.ReadRequest(r, &ml); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := ml.Verify(
			metrics.VerifyID,
			metrics.VerifyType,
			metrics.VerifyDelta,
			metrics.VerifyValue,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.service.BatchUpdate(r.Context(), ml)
		if err != nil {
			http.Error(w, server.ErrInternal.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
