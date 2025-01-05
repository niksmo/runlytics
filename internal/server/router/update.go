package router

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
)

const (
	gauge   = server.Gauge
	counter = server.Counter
)

type updateHandler struct {
	repo server.Repository
}

func SetUpdateRoute(r *chi.Mux, repo server.Repository) {
	h := &updateHandler{repo}
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", h.update())
		log.Println("Register endpoint: /update/{type}/{name}/{value}")
	})
}

func (h *updateHandler) update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.EscapedPath())

		t := server.MetricType(chi.URLParam(r, "type"))
		n := chi.URLParam(r, "name")
		v := chi.URLParam(r, "value")

		switch t {
		case counter:
			cV, err := strconv.ParseInt(v, 10, 64)
			if isErr(err, w) {
				return
			}
			h.repo.AddCounter(n, cV)
		case gauge:
			gV, err := strconv.ParseFloat(v, 64)
			if isErr(err, w) {
				return
			}
			h.repo.AddGauge(n, gV)
		default:
			err := errors.New("unexpected metrics type")
			if isErr(err, w) {
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func isErr(err error, w http.ResponseWriter) bool {
	ret := false

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		ret = true
		return ret
	}

	return ret
}
