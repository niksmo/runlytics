package storage

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

type memStorage struct {
	mu      sync.Mutex
	counter map[string]int64
	gauge   map[string]float64
}

func NewMemStorage() *memStorage {
	return &memStorage{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (ms *memStorage) SetCounter(name string, value int64) {
	ms.mu.Lock()
	prev := ms.counter[name]
	current := prev + value
	ms.counter[name] = current
	ms.mu.Unlock()

	logger.Log.Debug(
		"Set count metric",
		zap.String("name", name),
		zap.Int64("prev", prev),
		zap.Int64("current", current),
	)
}

func (ms *memStorage) SetGauge(name string, value float64) {
	ms.mu.Lock()
	prev := ms.gauge[name]
	ms.gauge[name] = value
	ms.mu.Unlock()

	logger.Log.Debug(
		"Set gauge metric",
		zap.String("name", name),
		zap.Float64("prev", prev),
		zap.Float64("current", value),
	)

}

func (ms *memStorage) GetCounter(name string) (int64, error) {
	ms.mu.Lock()
	value, ok := ms.counter[name]
	ms.mu.Unlock()

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	}
	return value, nil
}

func (ms *memStorage) GetGauge(name string) (float64, error) {
	ms.mu.Lock()
	value, ok := ms.gauge[name]
	ms.mu.Unlock()

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	}
	return value, nil
}

func (ms *memStorage) GetData() (gauge map[string]float64, counter map[string]int64) {
	gauge = make(map[string]float64, len(ms.gauge))
	counter = make(map[string]int64, len(ms.counter))

	ms.mu.Lock()
	for k, v := range ms.gauge {
		gauge[k] = v
	}

	for k, v := range ms.counter {
		counter[k] = v
	}
	ms.mu.Unlock()

	return
}
