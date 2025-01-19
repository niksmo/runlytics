package emitter

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
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
		sValue := strconv.FormatFloat(value, 'f', -1, 64)
		reqURL := makeReqURL(e.baseURL, "gauge", name, sValue)

		e.post(reqURL)
	}
}

func (e *HTTPEmitter) emitCounter() {
	logger.Log.Debug("Emit gauge metrics")
	for name, value := range e.metricsData.GetCounterMetrics() {
		sValue := strconv.FormatInt(value, 10)
		reqURL := makeReqURL(e.baseURL, "counter", name, sValue)

		e.post(reqURL)
	}
}

func (e *HTTPEmitter) post(reqURL string) {
	logger.Log.Info(
		"Start request",
		zap.String("URL", reqURL),
		zap.String("method", "POST"),
	)

	start := time.Now()
	res, err := e.client.Post(reqURL, "text/plain", http.NoBody)
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

	logger.Log.Info(
		"Got response",
		zap.String("URL", reqURL),
		zap.String("method", "POST"),
		zap.Duration("duration", time.Since(start)),
		zap.Int("statusCode", res.StatusCode),
	)

}

func makeReqURL(
	baseURL *url.URL,
	mType string,
	mName string,
	mValue string,
) string {
	return baseURL.JoinPath("update", mType, mName, mValue).String()
}
