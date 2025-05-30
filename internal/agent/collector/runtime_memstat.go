package collector

import (
	"runtime"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type RuntimeMemStat struct {
	data metricsData
	poll time.Duration
	mu   sync.RWMutex
}

func NewRuntimeMemStat(interval time.Duration) *RuntimeMemStat {
	collector := &RuntimeMemStat{
		poll: interval,
		data: newMetricsData(),
	}

	return collector
}

func (collector *RuntimeMemStat) Run() error {
	logger.Log.Info(
		"Run RuntimeMemStat collector",
		zap.Float64("interval", collector.poll.Seconds()),
	)

	for {
		start := time.Now()
		collector.collectMetrics()
		logger.Log.Debug(
			"RuntimeMemStat collect metrics",
			zap.Duration("duration", time.Since(start)),
		)
		time.Sleep(collector.poll)
	}
}

func (collector *RuntimeMemStat) GetGaugeMetrics() map[string]float64 {
	ret := make(map[string]float64, len(collector.data.gauge))
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	for k, v := range collector.data.gauge {
		ret[k] = v
	}

	return ret
}

func (collector *RuntimeMemStat) GetCounterMetrics() map[string]int64 {
	return collector.data.counter
}

func (collector *RuntimeMemStat) collectMetrics() {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	collector.mu.Lock()
	defer collector.mu.Unlock()

	collector.data.gauge["Alloc"] = float64(memStat.Alloc)
	collector.data.gauge["BuckHashSys"] = float64(memStat.BuckHashSys)
	collector.data.gauge["Frees"] = float64(memStat.Frees)
	collector.data.gauge["GCCPUFraction"] = memStat.GCCPUFraction
	collector.data.gauge["GCSys"] = float64(memStat.GCSys)
	collector.data.gauge["HeapAlloc"] = float64(memStat.HeapAlloc)
	collector.data.gauge["HeapIdle"] = float64(memStat.HeapIdle)
	collector.data.gauge["HeapInuse"] = float64(memStat.HeapInuse)
	collector.data.gauge["HeapObjects"] = float64(memStat.HeapObjects)
	collector.data.gauge["HeapReleased"] = float64(memStat.HeapReleased)
	collector.data.gauge["HeapSys"] = float64(memStat.HeapSys)
	collector.data.gauge["LastGC"] = float64(memStat.LastGC)
	collector.data.gauge["Lookups"] = float64(memStat.Lookups)
	collector.data.gauge["MCacheInuse"] = float64(memStat.MCacheInuse)
	collector.data.gauge["MCacheSys"] = float64(memStat.MCacheSys)
	collector.data.gauge["MSpanInuse"] = float64(memStat.MSpanInuse)
	collector.data.gauge["MSpanSys"] = float64(memStat.MSpanSys)
	collector.data.gauge["Mallocs"] = float64(memStat.Mallocs)
	collector.data.gauge["NextGC"] = float64(memStat.NextGC)
	collector.data.gauge["NumForcedGC"] = float64(memStat.NumForcedGC)
	collector.data.gauge["NumGC"] = float64(memStat.NumGC)
	collector.data.gauge["OtherSys"] = float64(memStat.OtherSys)
	collector.data.gauge["PauseTotalNs"] = float64(memStat.PauseTotalNs)
	collector.data.gauge["StackInuse"] = float64(memStat.StackInuse)
	collector.data.gauge["StackSys"] = float64(memStat.StackSys)
	collector.data.gauge["Sys"] = float64(memStat.Sys)
	collector.data.gauge["TotalAlloc"] = float64(memStat.TotalAlloc)
}
