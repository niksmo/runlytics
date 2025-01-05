package router

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server"
)

func SetUpdateRoute(r *chi.Mux, repo server.RepositoryUpdate) {
	h := &updateHandler{repo}
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", h.postHadleFunc())
		log.Println("Register endpoint: /update/{type}/{name}/{value}")
	})
}

type updateHandler struct {
	repo server.RepositoryUpdate
}

func (h *updateHandler) postHadleFunc() http.HandlerFunc {
	isErr := func(err error, w http.ResponseWriter) bool {
		ret := false

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			ret = true
			return ret
		}

		return ret
	}

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
			h.repo.SetGauge(n, gV)
		default:
			err := errors.New("unexpected metrics type")
			if isErr(err, w) {
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
