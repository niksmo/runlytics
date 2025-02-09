package emitter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

var triesDuration = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type HTTPEmitter struct {
	interval        time.Duration
	metricsData     di.GaugeCounterMetricsGetter
	client          *http.Client
	baseURL         *url.URL
	prevPollCounter int64
}

func New(
	interval time.Duration,
	metricsData di.GaugeCounterMetricsGetter,
	client *http.Client,
	baseURL *url.URL,
) *HTTPEmitter {
	return &HTTPEmitter{
		interval:        interval,
		metricsData:     metricsData,
		client:          client,
		baseURL:         baseURL,
		prevPollCounter: 0,
	}
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
		e.emit()
	}
}

func (e *HTTPEmitter) emit() {
	var batch metrics.MetricsBatchUpdate

	for name, value := range e.metricsData.GetGaugeMetrics() {
		batch = append(
			batch,
			metrics.MetricsUpdate{
				ID: name, MType: metrics.MTypeGauge, Value: &value,
			},
		)
	}

	for name, value := range e.metricsData.GetCounterMetrics() {
		if name == "PollCount" {
			prev := e.prevPollCounter
			e.prevPollCounter = value
			value = value - prev
		}
		batch = append(
			batch,
			metrics.MetricsUpdate{
				ID: name, MType: metrics.MTypeCounter, Delta: &value,
			},
		)

	}

	e.post(batch)
}

func (e *HTTPEmitter) post(metrics metrics.MetricsBatchUpdate) {
	reqURL := e.baseURL.JoinPath("updates").String()
	logger.Log.Info(
		"Start request",
		zap.String("URL", reqURL),
		zap.String("method", "POST"),
	)

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if err := json.NewEncoder(gzipWriter).Encode(metrics); err != nil {
		logger.Log.Debug("Encode to JSON error", zap.Error(err))
	}
	gzipWriter.Close()

	request, err := http.NewRequest("POST", reqURL, &buf)
	if err != nil {
		logger.Log.Warn("Error while creating http request", zap.Error(err))
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	start := time.Now()
	res, err := e.client.Do(request)
	if err != nil {
		for _, wait := range triesDuration {
			logger.Log.Debug("Do request", zap.Duration("try_after", wait), zap.Error(err))
			time.Sleep(wait)
			res, err = e.client.Do(request)
			if err == nil {
				break
			}
		}

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
	}
	defer res.Body.Close()

	var data []byte

	if res.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(res.Body)

		if err != nil {
			logger.Log.Warn("Error while creating new gzip reader", zap.Error(err))
		}
		res.Body = gzipReader
	}

	data, err = io.ReadAll(res.Body)

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
