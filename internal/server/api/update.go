package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/schemas"
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
		postPath := "/"
		r.Post(postPath, handler.post())
		debugLogRegister(path + postPath)
	})
}

func (handler *UpdateHandler) post() http.HandlerFunc {
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

		if err := handler.service.Update(&metrics); err != nil {
			writeTextErrorResponse(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		writeJSONResponse(w, metrics)
	}
}
