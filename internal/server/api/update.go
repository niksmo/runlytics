package api

import (
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

type UpdateHandler struct {
	service   di.UpdateService
	validator di.MetricsParamsSchemeVerifier
}

func SetUpdateHandler(
	mux *chi.Mux,
	service di.UpdateService,
	validator di.MetricsParamsSchemeVerifier,
) {
	path := "/update"
	handler := &UpdateHandler{service, validator}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.With(middleware.AllowJSON).Post(byJSONPath, handler.updateByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}/{value}"
		r.Post(byURLParamsPath, handler.updataByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

func (handler *UpdateHandler) updateByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var scheme metrics.MetricsUpdate
		if err := jsonhttp.ReadRequest(r, &scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := handler.validator.VerifyScheme(scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := handler.service.Update(r.Context(), &scheme.Metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = jsonhttp.WriteResponse(w, http.StatusOK, scheme)
		if err != nil {
			logger.Log.Error("error on write response", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *UpdateHandler) updataByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheme, err := handler.validator.VerifyParams(
			chi.URLParam(r, "name"),
			chi.URLParam(r, "type"),
			chi.URLParam(r, "value"),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = handler.service.Update(r.Context(), &scheme)
		if err != nil {
			http.Error(w, server.ErrInternal.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err = io.WriteString(w, scheme.GetValue()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
