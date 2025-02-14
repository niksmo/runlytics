package api

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type UpdateHandler struct {
	service   di.UpdateService
	validator di.MetricsParamsSchemeVerifier
}

func SetUpdateHandler(
	mux *chi.Mux,
	service di.UpdateService,
	validator di.MetricsParamsSchemeVerifier,
	config di.KeyGetter,
) {
	path := "/update"
	handler := &UpdateHandler{service, validator}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.With(
			middleware.AllowJSON,
			middleware.VerifyAndWriteSHA256(config.Key()),
		).Post(byJSONPath, handler.updateByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}/{value}"
		r.Post(byURLParamsPath, handler.updataByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

func (handler *UpdateHandler) updateByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var scheme metrics.MetricsUpdate
		if err := decodeJSON(r, &scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := handler.validator.VerifyScheme(&scheme); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resScheme, err := handler.service.Update(r.Context(), &scheme)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, resScheme)
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

		resScheme, err := handler.service.Update(r.Context(), scheme)
		if err != nil {
			http.Error(w, server.ErrInternal.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err = io.WriteString(w, resScheme.StrconvValue()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
