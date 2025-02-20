package emitter

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

const headerHashKey = "HashSHA256"

var waitIntervals = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type HTTPEmitter struct {
	interval            time.Duration
	metricsCollector    di.MetricsCollector
	client              *http.Client
	baseURL             *url.URL
	prevPollCounter     int64
	key                 string
	rollbackPollCounter func()
}

func New(
	config di.AgentConfig,
	metricsCollector di.MetricsCollector,
	client *http.Client,
) *HTTPEmitter {
	return &HTTPEmitter{
		interval:         config.Report(),
		metricsCollector: metricsCollector,
		client:           client,
		baseURL:          config.Addr(),
		prevPollCounter:  0,
		key:              config.Key(),
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

	for name, value := range e.metricsCollector.GetGaugeMetrics() {
		batch = append(
			batch,
			metrics.MetricsUpdate{
				ID: name, MType: metrics.MTypeGauge, Value: &value,
			},
		)
	}

	for name, value := range e.metricsCollector.GetCounterMetrics() {
		if name == "PollCount" {
			prev := e.prevPollCounter
			e.prevPollCounter = value
			value = value - prev
			e.rollbackPollCounter = func() {
				e.prevPollCounter = prev
			}
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

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Panic("Encode to JSON error", zap.Error(err))
	}

	var hexSHA256 string
	if e.key != "" {
		hexSHA256 = getHexHashSHA256(jsonData, e.key)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err = gzipWriter.Write(jsonData); err != nil {
		logger.Log.Panic("Write gzip", zap.Error(err))
	}
	gzipWriter.Close()

	request, err := http.NewRequest("POST", reqURL, &buf)
	if err != nil {
		logger.Log.Panic("Error while creating http request", zap.Error(err))
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")
	if hexSHA256 != "" {
		request.Header.Set(headerHashKey, hexSHA256)
	}

	start := time.Now()
	res, err := doRequestWithRetries(e.client, request)
	if err != nil {
		e.rollbackPollCounter()
		logger.Log.Info("Rollback poll counter", zap.Int64("current", e.prevPollCounter))
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

	if res.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(res.Body)

		if err != nil {
			logger.Log.Warn("Error while creating new gzip reader", zap.Error(err))
		}
		res.Body = gzipReader
	}

	var data []byte
	data, err = io.ReadAll(res.Body)

	if err != nil {
		logger.Log.Error("Read response data", zap.Error(err))
	}

	logger.Log.Info(
		"Got response",
		zap.String("URL", reqURL),
		zap.String("method", "POST"),
		zap.Duration("duration", time.Since(start)),
		zap.Int("statusCode", res.StatusCode),
		zap.String("hash", res.Header.Get(headerHashKey)),
		zap.String("data", string(data)),
	)

}

func doRequestWithRetries(
	client *http.Client, req *http.Request,
) (*http.Response, error) {
	res, err := client.Do(req)
	if err != nil {
		for _, interval := range waitIntervals {
			logger.Log.Debug("Do request", zap.Duration("tryAfter", interval), zap.Error(err))
			time.Sleep(interval)
			res, err = client.Do(req)
			if err == nil {
				break
			}
		}
	}
	return res, err
}

func getHexHashSHA256(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(data)
	if err != nil {
		logger.Log.Panic("Write to Hash", zap.Error(err))
	}
	return hex.EncodeToString(h.Sum(nil))
}
