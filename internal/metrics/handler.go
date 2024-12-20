package metrics

import (
	"log"
	"net/http"
	"strconv"
)

type metric string

const (
	counter metric = "counter"
	gauge   metric = "gauge"
)

type Storage interface {
	AddCount(name string, value int64)
	AddGauge(name string, value float64)
}

type MetricsHandler struct {
	storage Storage
}

func NewHandler(mux *http.ServeMux, storage Storage) {
	h := &MetricsHandler{storage}
	mux.HandleFunc(`POST /update/{type}/{name}/{value}`, h.Update())
	log.Println("Register endpoint: /update/{type}/{name}/{value}")
}

func (h *MetricsHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.EscapedPath())

		t := metric(r.PathValue("type"))
		n := r.PathValue("name")
		v := r.PathValue("value")

		switch t {
		case counter:
			cV, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			h.storage.AddCount(n, cV)
		case gauge:
			gV, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			h.storage.AddGauge(n, gV)
		default:
			log.Println("Unexpected type")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
