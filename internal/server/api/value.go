package api

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/metrics"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

type ReadHandler struct {
	service ReadService
}

type ReadService interface {
	Read(mData *metrics.Metrics) error
}

func SetReadHandler(mux *chi.Mux, service ReadService) {
	path := "/value"
	handler := &ReadHandler{service}
	mux.Route(path, func(r chi.Router) {
		byJSONPath := "/"
		r.Post(byJSONPath, handler.readByJSON())
		debugLogRegister(path + byJSONPath)

		byURLParamsPath := "/{type}/{name}"
		r.Get(byURLParamsPath, handler.readByURLParams())
		debugLogRegister(path + byURLParamsPath)
	})
}

func (handler *ReadHandler) readByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := verifyContentType(r, JSONMediaType); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusUnsupportedMediaType,
				err.Error(),
			)
			return
		}

		var metrics metrics.Metrics
		if err := decodeJSON(r, &metrics); err != nil {
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

		if err := handler.service.Read(&metrics); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusNotFound,
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

func (handler *ReadHandler) readByURLParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mData := metrics.Metrics{
			ID:    chi.URLParam(r, "name"),
			MType: chi.URLParam(r, "type"),
		}

		if err := handler.service.Read(&mData); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusNotFound,
				err.Error(),
			)
			return
		}

		resData, err := mData.StrconvValue()
		if err != nil {
			writeTextErrorResponse(
				w, http.StatusInternalServerError, server.ErrInternal.Error(),
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, resData)
	}
}
