package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"
)

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

type Repository interface {
	AddCounter(name string, value int64)
	AddGauge(name string, value float64)
}

type MetricsHandler struct {
	repo Repository
}

func NewHandler(router *http.ServeMux, repo Repository) {
	h := &MetricsHandler{repo}
	router.HandleFunc(`POST /update/{type}/{name}/{value}`, h.Update())
	log.Println("Register endpoint: /update/{type}/{name}/{value}")
}

func (h *MetricsHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.EscapedPath())

		t := MetricType(r.PathValue("type"))
		n := r.PathValue("name")
		v := r.PathValue("value")

		switch t {
		case Counter:
			cV, err := strconv.ParseInt(v, 10, 64)
			if isErr(err, w) {
				return
			}
			h.repo.AddCounter(n, cV)
		case Gauge:
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
