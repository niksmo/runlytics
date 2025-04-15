package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/jsonhttp"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type ValueHandler struct {
	service   di.ReadService
	validator di.MetricsParamsSchemeVerifier
}

func SetValueHandler(
	mux *chi.Mux,
	service di.ReadService,
	validator di.MetricsParamsSchemeVerifier,
) {
	path := "/value"
	handler := &ValueHandler{service, validator}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.With(middleware.AllowJSON).Post(byJSONPath, handler.readByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}"
		r.Get(byURLParamsPath, handler.readByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

func (handler *ValueHandler) readByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var scheme metrics.MetricsRead
		if err := jsonhttp.ReadRequest(r, &scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := handler.validator.VerifyScheme(scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := handler.service.Read(r.Context(), &scheme.Metrics)
		if err != nil {
			if errors.Is(err, server.ErrNotExists) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(
				w, server.ErrInternal.Error(), http.StatusInternalServerError,
			)
			return
		}

		err = jsonhttp.WriteResponse(w, http.StatusOK, scheme)
		if err != nil {
			logger.Log.Error("error on write response", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *ValueHandler) readByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheme, err := handler.validator.VerifyParams(
			chi.URLParam(r, "name"),
			chi.URLParam(r, "type"),
			"skipped",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = handler.service.Read(r.Context(), &scheme)
		if err != nil {
			if errors.Is(err, server.ErrNotExists) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(
				w, server.ErrInternal.Error(), http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err = io.WriteString(w, scheme.GetValue()); err != nil {
			http.Error(
				w, err.Error(), http.StatusInternalServerError,
			)
		}
	}
}
