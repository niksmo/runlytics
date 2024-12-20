package handler

import (
	"fmt"
	"log"
	"net/http"
)

type Storage interface {
	AddCount(name string, value int)
	AddGauge(name string, value float64)
}

func NewUpdate(mux *http.ServeMux) {
	mux.HandleFunc(`POST /update/{type}/{name}/{value}`, updateMetrics)
	log.Println("Register endpoint: /update/{type}/{name}/{value}")
}

func updateMetrics(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.EscapedPath())

	t := r.PathValue("type")
	n := r.PathValue("name")
	v := r.PathValue("value")
	w.Write([]byte(fmt.Sprintf("Type=%s name=%s value=%s\n", t, n, v)))
}
