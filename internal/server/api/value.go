package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/schemas"
)

type ValueHandler struct {
	service ValueService
}

type ValueService interface {
	Read(metrics *schemas.Metrics) error
}

func SetReadHandler(mux *chi.Mux, service ValueService) {
	path := "/value"
	handler := &ValueHandler{service}
	mux.Route(path, func(r chi.Router) {
		postPath := "/"
		r.Post(postPath, handler.post())
		debugLogRegister(path + postPath)
	})
}

func (handler *ValueHandler) post() http.HandlerFunc {
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

		if err := handler.service.Read(&metrics); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusNotFound,
				err.Error(),
			)
			return
		}

		writeJSONResponse(w, metrics)
	}
}
