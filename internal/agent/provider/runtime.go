package provider

import (
	"runtime"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type runtimeStat struct {
	data   metricsData
	poll   time.Duration
	mu     sync.RWMutex
	ticker *time.Ticker
}

func newRuntimeStat(poll time.Duration) *runtimeStat {
	return &runtimeStat{
		poll:   poll,
		data:   newMetricsData(),
		ticker: time.NewTicker(poll),
	}
}

func (s *runtimeStat) Run() {
	const op = "runtimeStat.Run"
	log := logger.Log.With(
		zap.String("op", op), zap.Duration("updateInt", s.poll),
	)
	log.Info("running")

	for range s.ticker.C {
		s.updateData()
		log.Debug("update data")
	}
}

func (s *runtimeStat) Stop() {
	const op = "runtimeStat.Stop"
	s.ticker.Stop()
	logger.Log.Info("stopped", zap.String("op", op))
}

func (s *runtimeStat) GetMetrics() metrics.MetricsList {
	s.mu.RLock()
	defer s.mu.RUnlock()
	gData := s.readGauge()
	cData := s.readCounter()

	mSize := len(gData) + len(cData)
	m := make(metrics.MetricsList, 0, mSize)

	for n, v := range gData {
		m = append(
			m, metrics.Metrics{ID: n, Value: v, MType: metrics.MTypeGauge},
		)
	}

	for n, v := range cData {
		m = append(
			m, metrics.Metrics{ID: n, Delta: v, MType: metrics.MTypeCounter},
		)
	}

	return m
}

func (s *runtimeStat) readGauge() map[string]float64 {
	return s.data.gauge
}

func (s *runtimeStat) readCounter() map[string]int64 {
	return s.data.counter
}

func (s *runtimeStat) updateData() {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.gauge["Alloc"] = float64(memStat.Alloc)
	s.data.gauge["BuckHashSys"] = float64(memStat.BuckHashSys)
	s.data.gauge["Frees"] = float64(memStat.Frees)
	s.data.gauge["GCCPUFraction"] = memStat.GCCPUFraction
	s.data.gauge["GCSys"] = float64(memStat.GCSys)
	s.data.gauge["HeapAlloc"] = float64(memStat.HeapAlloc)
	s.data.gauge["HeapIdle"] = float64(memStat.HeapIdle)
	s.data.gauge["HeapInuse"] = float64(memStat.HeapInuse)
	s.data.gauge["HeapObjects"] = float64(memStat.HeapObjects)
	s.data.gauge["HeapReleased"] = float64(memStat.HeapReleased)
	s.data.gauge["HeapSys"] = float64(memStat.HeapSys)
	s.data.gauge["LastGC"] = float64(memStat.LastGC)
	s.data.gauge["Lookups"] = float64(memStat.Lookups)
	s.data.gauge["MCacheInuse"] = float64(memStat.MCacheInuse)
	s.data.gauge["MCacheSys"] = float64(memStat.MCacheSys)
	s.data.gauge["MSpanInuse"] = float64(memStat.MSpanInuse)
	s.data.gauge["MSpanSys"] = float64(memStat.MSpanSys)
	s.data.gauge["Mallocs"] = float64(memStat.Mallocs)
	s.data.gauge["NextGC"] = float64(memStat.NextGC)
	s.data.gauge["NumForcedGC"] = float64(memStat.NumForcedGC)
	s.data.gauge["NumGC"] = float64(memStat.NumGC)
	s.data.gauge["OtherSys"] = float64(memStat.OtherSys)
	s.data.gauge["PauseTotalNs"] = float64(memStat.PauseTotalNs)
	s.data.gauge["StackInuse"] = float64(memStat.StackInuse)
	s.data.gauge["StackSys"] = float64(memStat.StackSys)
	s.data.gauge["Sys"] = float64(memStat.Sys)
	s.data.gauge["TotalAlloc"] = float64(memStat.TotalAlloc)
}
