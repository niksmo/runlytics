package collector

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/counter"
	"go.uber.org/zap"
)

type ManualStat struct {
	poll    time.Duration
	data    metricsData
	counter *counter.Counter
	mu      sync.RWMutex
}

func NewManualStat(interval time.Duration) *ManualStat {
	collector := &ManualStat{
		poll:    interval,
		data:    newMetricsData(),
		counter: counter.New(),
	}

	return collector
}

func (collector *ManualStat) Run() {
	logger.Log.Info(
		"Run ManualStat collector", zap.Float64("interval", collector.poll.Seconds()),
	)

	for {
		start := time.Now()
		collector.collectMetrics()
		logger.Log.Debug(
			"ManualStat collect metrics",
			zap.Duration("duration", time.Since(start)),
		)
		time.Sleep(collector.poll)
	}
}

func (collector *ManualStat) GetGaugeMetrics() map[string]float64 {
	ret := make(map[string]float64, len(collector.data.gauge))
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	for k, v := range collector.data.gauge {
		ret[k] = v
	}

	return ret
}

func (collector *ManualStat) GetCounterMetrics() map[string]int64 {
	ret := make(map[string]int64, len(collector.data.counter))
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	for k, v := range collector.data.counter {
		ret[k] = v
	}

	return ret
}

func (collector *ManualStat) collectMetrics() {
	collector.mu.Lock()
	defer collector.mu.Unlock()
	collector.data.gauge["RandomValue"] = rand.Float64()
	collector.data.counter["PollCount"] = collector.counter.Next()
}
