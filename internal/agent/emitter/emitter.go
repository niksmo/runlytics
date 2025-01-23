package emitter

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/schemas"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

type HTTPClient interface {
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type MetricsData interface {
	GetGaugeMetrics() map[string]float64
	GetCounterMetrics() map[string]int64
}

type HTTPEmitter struct {
	interval        time.Duration
	metricsData     MetricsData
	client          HTTPClient
	baseURL         *url.URL
	prevPollCounter int64
}

func New(
	interval time.Duration,
	metricsData MetricsData,
	client HTTPClient,
	baseURL *url.URL,
) *HTTPEmitter {

	return &HTTPEmitter{interval, metricsData, client, baseURL, int64(0)}
}

func (e *HTTPEmitter) Run() {
	logger.Log.Info(
		"Run HTTPEmitter",
		zap.String("addr", e.baseURL.String()),
		zap.Float64("interval", e.interval.Seconds()),
	)

	for {
		logger.Log.Debug("Wait", zap.Float64("seconds", e.interval.Seconds()))
		time.Sleep(e.interval)
		e.emitGauge()
		e.emitCounter()
	}
}

func (e *HTTPEmitter) emitGauge() {
	logger.Log.Debug("Emit gauge metrics")
	for name, value := range e.metricsData.GetGaugeMetrics() {
		gaugeMetrics := schemas.Metrics{ID: name, MType: server.MTypeGauge, Value: &value}
		e.post(gaugeMetrics)
	}
}

func (e *HTTPEmitter) emitCounter() {
	logger.Log.Debug("Emit counter metrics")
	for name, value := range e.metricsData.GetCounterMetrics() {
		delta := value - e.prevPollCounter
		e.prevPollCounter = value
		counterMetrics := schemas.Metrics{ID: name, MType: server.MTypeCounter, Delta: &delta}

		e.post(counterMetrics)
	}
}

func (e *HTTPEmitter) post(metrics schemas.Metrics) {
	reqURL := e.baseURL.JoinPath("update").String()
	logger.Log.Info(
		"Start request",
		zap.String("URL", reqURL),
		zap.String("method", "POST"),
	)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metrics); err != nil {
		logger.Log.Debug("Encode to JSON error", zap.Error(err))
	}

	start := time.Now()
	res, err := e.client.Post(reqURL, "application/json", &buf)
	if err != nil {
		logger.Log.Info(
			"Got response",
			zap.String("URL", reqURL),
			zap.String("method", "POST"),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Log.Error("Read response data error", zap.Error(err))
	}

	logger.Log.Info(
		"Got response",
		zap.String("URL", reqURL),
		zap.String("method", "POST"),
		zap.Duration("duration", time.Since(start)),
		zap.Int("statusCode", res.StatusCode),
		zap.String("data", string(data)),
	)

}
