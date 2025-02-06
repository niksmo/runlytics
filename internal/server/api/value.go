package api

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/metrics"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/validator"
)

type ValueHandler struct {
	service   ValueService
	validator ValueValidator
}

type ValueService interface {
	Read(ctx context.Context, mData *metrics.MetricsRead) (metrics.Metrics, error)
}

type ValueValidator interface {
	VerifyScheme(validator.Verifier) error
}

func SetValueHandler(
	mux *chi.Mux, service ValueService, validator ValueValidator,
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
		if err := decodeJSON(r, &scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := handler.validator.VerifyScheme(&scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resScheme, err := handler.service.Read(r.Context(), &scheme)
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

		writeJSONResponse(w, resScheme)
	}
}

func (handler *ValueHandler) readByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheme := &metrics.MetricsRead{
			ID:    chi.URLParam(r, "name"),
			MType: chi.URLParam(r, "type"),
		}

		if err := handler.validator.VerifyScheme(scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resScheme, err := handler.service.Read(r.Context(), scheme)
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
		if _, err = io.WriteString(w, resScheme.StrconvValue()); err != nil {
			http.Error(
				w, err.Error(), http.StatusInternalServerError,
			)
		}
	}
}
