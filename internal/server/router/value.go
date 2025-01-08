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
	isErr := func(err error, w http.ResponseWriter) bool {
		ret := false

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			ret = true
			return ret
		}

		return ret
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.EscapedPath())

		t := server.MetricType(chi.URLParam(r, "type"))
		n := chi.URLParam(r, "name")
		var v string

		switch t {
		case counter:
			cV, err := h.repo.GetCounter(n)
			if isErr(err, w) {
				return
			}
			v = strconv.FormatInt(cV, 10)
		case gauge:
			gV, err := h.repo.GetGauge(n)
			if isErr(err, w) {
				return
			}
			v = strconv.FormatFloat(gV, 'f', -1, 64)
		default:
			err := errors.New("unexpected metrics type")
			if isErr(err, w) {
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, v)
	}

}
