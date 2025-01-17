package storage

import (
	"errors"
	"fmt"
	"log"
	"sync"
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
	log.Printf(
		"Add count metric: name=%q prev=%v current=%v\n",
		name, prev, current,
	)
}

func (ms *memStorage) SetGauge(name string, value float64) {
	ms.mu.Lock()
	prev := ms.gauge[name]
	ms.gauge[name] = value
	ms.mu.Unlock()
	log.Printf(
		"Add gauge metric: name=%q prev=%v current=%v\n",
		name, prev, value,
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
