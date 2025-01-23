package repository

import (
	"errors"
	"fmt"
	"sync"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

var (
	ErrNotExists = errors.New("not exists")
)

type repository struct {
	mu      sync.Mutex
	counter map[string]int64
	gauge   map[string]float64
}

func New() *repository {
	return &repository{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (ms *repository) UpdateCounterByName(name string, value int64) int64 {
	ms.mu.Lock()
	prev := ms.counter[name]
	current := prev + value
	ms.counter[name] = current
	ms.mu.Unlock()

	logger.Log.Debug(
		"Update count metric",
		zap.String("name", name),
		zap.Int64("prev", prev),
		zap.Int64("current", current),
	)

	return current
}

func (ms *repository) UpdateGaugeByName(name string, value float64) float64 {
	ms.mu.Lock()
	prev := ms.gauge[name]
	ms.gauge[name] = value
	ms.mu.Unlock()

	logger.Log.Debug(
		"Update gauge metric",
		zap.String("name", name),
		zap.Float64("prev", prev),
		zap.Float64("current", value),
	)
	return value
}

func (ms *repository) ReadCounterByName(name string) (int64, error) {
	ms.mu.Lock()
	value, ok := ms.counter[name]
	ms.mu.Unlock()

	if !ok {
		logger.Log.Debug("Not found counter metric", zap.String("name", name))
		return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	}
	return value, nil
}

func (ms *repository) ReadGaugeByName(name string) (float64, error) {
	ms.mu.Lock()
	value, ok := ms.gauge[name]
	ms.mu.Unlock()

	if !ok {
		logger.Log.Debug("Not found gauge metric", zap.String("name", name))
		return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	}
	return value, nil
}

func (ms *repository) ReadGauge() map[string]float64 {
	gauge := make(map[string]float64, len(ms.gauge))

	ms.mu.Lock()
	for k, v := range ms.gauge {
		gauge[k] = v
	}
	ms.mu.Unlock()

	return gauge
}

func (ms *repository) ReadCounter() map[string]int64 {
	counter := make(map[string]int64, len(ms.counter))

	ms.mu.Lock()
	for k, v := range ms.counter {
		counter[k] = v
	}
	ms.mu.Unlock()

	return counter
}
