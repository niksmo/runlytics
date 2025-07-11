package httpapi

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/server/app/http/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

// UpdateHandler work with service and provides
// UpdateByJSON, UpdateByURLParams methods.
type UpdateHandler struct {
	service di.IUpdateService
}

// SetUpdateHandler sets UpdateHandler to "/update" path.
//
//   - "/update/" to UpdateByJSON method, only JSON media type is allowed
//   - "/update/{type}/{name}/{value}" to UpdateByURLParams method
func SetUpdateHandler(mux *chi.Mux, service di.IUpdateService) {
	path := "/update"
	handler := &UpdateHandler{service}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.With(middleware.AllowJSON).Post(byJSONPath, handler.UpdateByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}/{value}"
		r.Post(byURLParamsPath, handler.UpdataByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

// UpdateByJSON reads JSON data from request body.
func (h *UpdateHandler) UpdateByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m metrics.Metrics
		if err := ReadJSONRequest(r, &m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := checkMetricsForUpdate(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.service.Update(r.Context(), &m)
		if err != nil {
			http.Error(
				w, server.ErrInternal.Error(), http.StatusInternalServerError,
			)
			return
		}

		err = WriteJSONResponse(w, http.StatusOK, m)
		if err != nil {
			logger.Log.Error("error on write response", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// UpdataByURLParams reads data from URL params.
func (h *UpdateHandler) UpdataByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := metrics.NewFromStrArgs(
			chi.URLParam(r, "name"),
			chi.URLParam(r, "type"),
			chi.URLParam(r, "value"),
		)

		err := checkMetricsForUpdate(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.service.Update(r.Context(), &m)
		if err != nil {
			http.Error(
				w, server.ErrInternal.Error(), http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err = io.WriteString(w, m.GetValue()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func checkMetricsForUpdate(m metrics.Metrics) error {
	return m.Verify(
		metrics.VerifyID,
		metrics.VerifyType,
	)
}
