package provider

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/counter"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type manualStat struct {
	mu      sync.RWMutex
	counter *counter.Counter
	data    metricsData
	poll    time.Duration
	ticker  *time.Ticker
}

func newManualStat(poll time.Duration) *manualStat {
	return &manualStat{
		poll:    poll,
		data:    newMetricsData(),
		counter: counter.New(),
		ticker:  time.NewTicker(poll),
	}
}

// Run run manual statistics provider
func (s *manualStat) Run() {
	const op = "manualstat.Run"
	log := logger.Log.With(
		zap.String("op", op), zap.Duration("updateInt", s.poll),
	)
	log.Info("running")
	for range s.ticker.C {
		s.updateData()
		log.Debug("update data")
	}
}

func (s *manualStat) Stop() {
	const op = "manualstat.Stop"
	s.ticker.Stop()
	logger.Log.Info("stopped", zap.String("op", op))
}

func (s *manualStat) GetMetrics() metrics.MetricsList {
	s.mu.RLock()
	defer s.mu.RUnlock()
	gData := s.readGauge()
	cData := s.readCounter()

	mSize := len(gData) + len(cData)
	m := make(metrics.MetricsList, 0, mSize)

	for n, v := range gData {
		m = append(
			m, metrics.Metrics{ID: n, Value: &v, MType: metrics.MTypeGauge},
		)
	}

	for n, v := range cData {
		m = append(
			m, metrics.Metrics{ID: n, Delta: &v, MType: metrics.MTypeCounter},
		)
	}

	return m
}

func (s *manualStat) readGauge() map[string]float64 {
	return s.data.gauge
}

func (s *manualStat) readCounter() map[string]int64 {
	return s.data.counter
}

func (s *manualStat) updateData() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.gauge["RandomValue"] = rand.Float64()
	s.data.counter["PollCount"] = s.counter.Next()
}
