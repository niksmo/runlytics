package provider

import (
	"fmt"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type psUtilStat struct {
	data   metricsData
	poll   time.Duration
	mu     sync.RWMutex
	ticker *time.Ticker
}

func newPSUtilStat(poll time.Duration) *psUtilStat {
	return &psUtilStat{
		poll:   poll,
		data:   newMetricsData(),
		ticker: time.NewTicker(poll),
	}
}

func (s *psUtilStat) Run() {
	const op = "psutilstat.Run"
	log := logger.Log.With(
		zap.String("op", op), zap.Duration("updateInt", s.poll),
	)
	log.Info("running")

	for range s.ticker.C {
		if err := s.updateData(); err != nil {
			log.Warn("failed to update data", zap.Error(err))
		}
		log.Debug("update data")
	}
}

func (s *psUtilStat) Stop() {
	const op = "psutilstat.Stop"
	s.ticker.Stop()
	logger.Log.Info("stopped", zap.String("op", op))
}

func (s *psUtilStat) GetMetrics() metrics.MetricsList {
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

func (s *psUtilStat) readGauge() map[string]float64 {
	return s.data.gauge
}

func (s *psUtilStat) readCounter() map[string]int64 {
	return s.data.counter
}

func (s *psUtilStat) updateData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	virtualMemStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	s.data.gauge["TotalMemory"] = float64(virtualMemStat.Total)
	s.data.gauge["FreeMemory"] = float64(virtualMemStat.Free)

	cpuTimesStatList, err := cpu.Times(true)
	if err != nil {
		return err
	}

	for idx, cpuTimeStat := range cpuTimesStatList {
		name := fmt.Sprintf("CPUutilization%02v", idx+1)
		usageTime := cpuTimeStat.User + cpuTimeStat.System
		totalTime := usageTime + cpuTimeStat.Idle
		s.data.gauge[name] = usageTime / totalTime
	}
	return nil
}
