package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type PsUtilStat struct {
	data metricsData
	poll time.Duration
	mu   sync.RWMutex
}

func NewPsUtilStat(interval time.Duration) *PsUtilStat {
	collector := &PsUtilStat{
		poll: interval,
		data: newMetricsData(),
	}

	return collector
}

func (collector *PsUtilStat) Run() {
	logger.Log.Info(
		"Run PsUtilStat collector", zap.Float64("interval", collector.poll.Seconds()),
	)

	for {
		start := time.Now()
		collector.collectMetrics()
		logger.Log.Debug(
			"PsUtilStat collect metrics",
			zap.Duration("duration", time.Since(start)),
		)
		time.Sleep(collector.poll)
	}
}

func (collector *PsUtilStat) GetGaugeMetrics() map[string]float64 {
	ret := make(map[string]float64, len(collector.data.gauge))
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	for k, v := range collector.data.gauge {
		ret[k] = v
	}

	return ret
}

func (collector *PsUtilStat) GetCounterMetrics() map[string]int64 {
	return collector.data.counter
}

func (collector *PsUtilStat) collectMetrics() {
	collector.mu.Lock()
	defer collector.mu.Unlock()

	virtualMemStat, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Panic("Get VirtualMemoryStat")
	}
	collector.data.gauge["TotalMemory"] = float64(virtualMemStat.Total)
	collector.data.gauge["FreeMemory"] = float64(virtualMemStat.Free)

	cpuTimesStatList, err := cpu.Times(true)
	if err != nil {
		logger.Log.Panic("Get CPUTimesStat")
	}
	for idx, cpuTimeStat := range cpuTimesStatList {
		name := fmt.Sprintf("CPUutilization%02v", idx+1)
		usageTime := cpuTimeStat.User + cpuTimeStat.System
		totalTime := usageTime + cpuTimeStat.Idle
		collector.data.gauge[name] = usageTime / totalTime
	}
}
