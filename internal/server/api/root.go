package api

import (
	"bytes"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HTMLHandler struct {
	service HTMLService
}

type HTMLService interface {
	RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error
}

func SetHTMLHandler(mux *chi.Mux, service HTMLService) {
	path := "/"
	handler := &HTMLHandler{service}
	mux.Route(path, func(r chi.Router) {
		r.Get(path, handler.get())
		debugLogRegister(path)
	})
}

func (handler *HTMLHandler) get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		err := handler.service.RenderMetricsList(r.Context(), &buf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(ContentType, "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err = buf.WriteTo(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
