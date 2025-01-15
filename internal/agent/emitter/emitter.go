package emitter

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const contentType = "text/plain"
const updatePath = "update"

type HTTPClient interface {
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type MetricsData interface {
	GetGaugeMetrics() map[string]float64
	GetCounterMetrics() map[string]int64
}

type HTTPEmitter struct {
	interval    time.Duration
	metricsData MetricsData
	client      HTTPClient
	baseURL     *url.URL
}

func New(
	interval time.Duration,
	metricsData MetricsData,
	client HTTPClient,
	baseURL *url.URL,
) *HTTPEmitter {

	return &HTTPEmitter{interval, metricsData, client, baseURL}
}

func (e *HTTPEmitter) Run() {
	log.Printf(
		"Run HTTPEmitter with report interval = %vs\n",
		e.interval.Seconds(),
	)
	for {
		log.Print("Ready for emitting on host ", e.baseURL)
		log.Println("[REPORT]Wait for", e.interval.Seconds(), "sec")
		time.Sleep(e.interval)
		log.Println("[REPORT]Emit gauge metrics")
		e.emitGauge()
		log.Println("[REPORT]Emit counter metrics")
		e.emitCounter()
	}
}

func (e *HTTPEmitter) emitGauge() {
	for name, value := range e.metricsData.GetGaugeMetrics() {
		sValue := strconv.FormatFloat(value, 'f', -1, 64)
		reqURL := e.baseURL.JoinPath(updatePath, "gauge", name, sValue).String()

		e.post(reqURL)
	}
}

func (e *HTTPEmitter) emitCounter() {
	for name, value := range e.metricsData.GetCounterMetrics() {
		sValue := strconv.FormatInt(value, 10)
		reqURL := e.baseURL.JoinPath(updatePath, "counter", name, sValue).String()

		e.post(reqURL)
	}
}

func (e *HTTPEmitter) post(reqURL string) {
	log.Println("POST", reqURL, "start")
	res, err := e.client.Post(reqURL, contentType, http.NoBody)
	if err != nil {
		log.Println("POST", reqURL, "error:", err)
		return
	}
	defer res.Body.Close()

	log.Println("POST", reqURL, "response status:", res.StatusCode)

}
