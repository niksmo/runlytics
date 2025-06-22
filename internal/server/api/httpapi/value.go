package httpapi

import (
	"errors"
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

// ValueHandler works with service and provides
// ReadByJSON and ReadByURLParams methods.
type ValueHandler struct {
	service di.IReadService
}

// SetValueHandler sets ValueHandler to "/value" path.
//
//   - "/value/" to ReadByJSON method, only JSON media type is allowed
//   - "/value/{type}/{name}" to ReadByURLParams method
func SetValueHandler(mux *chi.Mux, service di.IReadService) {
	path := "/value"
	handler := &ValueHandler{service}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.With(middleware.AllowJSON).Post(byJSONPath, handler.ReadByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}"
		r.Get(byURLParamsPath, handler.ReadByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

// ReadByJSON reads JSON data from request body.
func (h *ValueHandler) ReadByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m metrics.Metrics
		if err := ReadJSONRequest(r, &m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := checkMetricsForRead(m)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.service.Read(r.Context(), &m)
		if errors.Is(err, server.ErrNotExists) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(
				w,
				server.ErrInternal.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		err = WriteJSONResponse(w, http.StatusOK, m)
		if err != nil {
			logger.Log.Error("error on write response", zap.Error(err))
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
		}
	}
}

// ReadByURLParams reads data from URL params.
func (h *ValueHandler) ReadByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := metrics.NewFromStrArgs(
			chi.URLParam(r, "name"),
			chi.URLParam(r, "type"),
			"",
		)
		err := checkMetricsForRead(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.service.Read(r.Context(), &m)
		if errors.Is(err, server.ErrNotExists) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(
				w,
				server.ErrInternal.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err = io.WriteString(w, m.GetValue()); err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
		}
	}
}

func checkMetricsForRead(m metrics.Metrics) error {
	return m.Verify(
		metrics.VerifyID,
		metrics.VerifyType,
	)
}
