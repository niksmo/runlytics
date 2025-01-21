package router

import (
	"errors"

	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

func SetUpdateRoute(r *chi.Mux, repo server.RepoUpdate) {
	h := &updateHandler{repo}
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", h.postHadleFunc())
		debugLogRegister("/update/{type}/{name}/{value}")
	})
}

type updateHandler struct {
	repo server.RepoUpdate
}

func (h *updateHandler) postHadleFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		t := server.MetricType(chi.URLParam(r, "type"))
		n := chi.URLParam(r, "name")
		v := chi.URLParam(r, "value")

		switch t {
		case counter:
			cV, err := strconv.ParseInt(v, 10, 64)
			if isUpdateErr(err, w) {
				return
			}
			h.repo.SetCounter(n, cV)
		case gauge:
			gV, err := strconv.ParseFloat(v, 64)
			if isUpdateErr(err, w) {
				return
			}
			h.repo.SetGauge(n, gV)
		default:
			err := errors.New("unexpected metrics type")
			if isUpdateErr(err, w) {
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func isUpdateErr(err error, w http.ResponseWriter) bool {

	if err != nil {
		logger.Log.Debug("Input metric error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	return false
}
