package httpapi

import (
	"bytes"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/pkg/di"
)

// HTMLHandler works with service and provide MetricsHTML method.
type HTMLHandler struct {
	service di.IHTMLService
}

// SetHTMLHandler sets MetricsHTML handler to "/" path.
func SetHTMLHandler(mux *chi.Mux, service di.IHTMLService) {
	path := "/"
	handler := &HTMLHandler{service}
	mux.Route(path, func(r chi.Router) {
		r.Get(path, handler.MetricsHTML())
		debugLogRegister(path)
	})
}

// MetricsHTML renders metrics list.
func (handler *HTMLHandler) MetricsHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		err := handler.service.RenderMetricsList(r.Context(), &buf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(ContentType, HTML)
		w.WriteHeader(http.StatusOK)
		if _, err = buf.WriteTo(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
