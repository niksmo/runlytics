package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/metrics"
)

type UpdateHandler struct {
	service UpdateService
}

type UpdateService interface {
	Update(mData *metrics.MetricsUpdate) error
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
				w, http.StatusUnsupportedMediaType, err.Error(),
			)
			return
		}

		var metrics metrics.MetricsUpdate
		if err := decodeJSON(r, &metrics); err != nil {
			writeTextErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := handler.service.Update(&metrics); err != nil {
			writeTextErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSONResponse(w, metrics)
	}
}

func (handler *UpdateHandler) updataByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mData := metrics.Metrics{
			ID:    chi.URLParam(r, "name"),
			MType: chi.URLParam(r, "type"),
		}
		sValue := chi.URLParam(r, "value")

		switch mData.MType {
		case metrics.MTypeCounter:
			delta, err := strconv.ParseInt(sValue, 10, 64)
			if err != nil {
				writeTextErrorResponse(
					w,
					http.StatusBadRequest,
					"counter value format should be 'integer'",
				)
				return
			}
			mData.Delta = &delta

		case metrics.MTypeGauge:
			value, err := strconv.ParseFloat(sValue, 64)
			if err != nil {
				writeTextErrorResponse(
					w,
					http.StatusBadRequest,
					"gauge value format should be 'float'",
				)
				return
			}
			mData.Value = &value
		}

		if err := handler.service.Update(&mData); err != nil {
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
