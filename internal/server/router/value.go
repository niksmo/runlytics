package router

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
)

func SetValueRoute(r *chi.Mux, repo server.RepoReadByName) {
	h := &valueHandler{repo}
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", h.getHandleFunc())
		log.Println("Register endpoint: /value/{type}/{name}")
	})
}

type valueHandler struct {
	repo server.RepoReadByName
}

func (h *valueHandler) getHandleFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.EscapedPath())

		t := server.MetricType(chi.URLParam(r, "type"))
		n := chi.URLParam(r, "name")
		var v string

		switch t {
		case counter:
			cV, err := h.repo.GetCounter(n)
			if isValueErr(err, w) {
				return
			}
			v = strconv.FormatInt(cV, 10)
		case gauge:
			gV, err := h.repo.GetGauge(n)
			if isValueErr(err, w) {
				return
			}
			v = strconv.FormatFloat(gV, 'f', -1, 64)
		default:
			err := errors.New("unexpected metrics type")
			if isValueErr(err, w) {
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, v)
	}

}

func isValueErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return true
	}

	return false
}
