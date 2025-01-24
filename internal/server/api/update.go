package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/schemas"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

type UpdateHandler struct {
	service UpdateService
}

type UpdateService interface {
	Update(metrics *schemas.Metrics) error
}

func SetUpdateHandler(mux *chi.Mux, service UpdateService) {
	path := "/update"
	handler := &UpdateHandler{service}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.Post(byJSONPath, handler.updateByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}/{value}"
		r.Post(byURLParamsPath, handler.updataByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

func (handler *UpdateHandler) updateByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := verifyContentType(r, JSONMediaType); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusUnsupportedMediaType,
				err.Error(),
			)
			return
		}

		var metrics schemas.Metrics
		if err := decodeJSONSchema(r, &metrics); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		logger.Log.Debug(
			"Decoded from JSON", zap.String("struct", metrics.String()),
		)

		if err := handler.service.Update(&metrics); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		logger.Log.Debug(
			"For encode to JSON", zap.String("struct", metrics.String()),
		)

		writeJSONResponse(w, metrics)
	}
}

func (handler *UpdateHandler) updataByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := schemas.Metrics{
			ID:    chi.URLParam(r, "name"),
			MType: chi.URLParam(r, "type"),
		}
		sValue := chi.URLParam(r, "value")

		switch metrics.MType {
		case server.MTypeCounter:
			delta, err := strconv.ParseInt(sValue, 10, 64)
			if err != nil {
				writeTextErrorResponse(
					w,
					http.StatusBadRequest,
					"counter value format, integer is expected",
				)
				return
			}
			metrics.Delta = &delta

		case server.MTypeGauge:
			value, err := strconv.ParseFloat(sValue, 64)
			if err != nil {
				writeTextErrorResponse(
					w,
					http.StatusBadRequest,
					"gauge value format, float is expected",
				)
				return
			}
			metrics.Value = &value
		}

		if err := handler.service.Update(&metrics); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
