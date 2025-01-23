package api

import (
	"bytes"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HTMLHandler struct {
	service HTMLService
}

type HTMLService interface {
	RenderMetricsList(buf *bytes.Buffer) error
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
		if err := handler.service.RenderMetricsList(&buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set(ContentTypePath, "text/html; charset=utf-8")
		buf.WriteTo(w)
	}
}
