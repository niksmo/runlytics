package storage

import (
	"log"
	"sync"
)

type memStorage struct {
	mu    sync.Mutex
	count map[string]int64
	gauge map[string]float64
}

func NewMemStorage() *memStorage {
	return &memStorage{
		count: make(map[string]int64),
		gauge: make(map[string]float64),
	}
}

func (ms *memStorage) AddCounter(name string, value int64) {
	ms.mu.Lock()
	current := ms.count[name] + value
	ms.count[name] = current
	ms.mu.Unlock()
	log.Printf(
		"Add count metric: name=%q value=%v currentValue=%v\n",
		name, value, current,
	)
}

func (ms *memStorage) AddGauge(name string, value float64) {
	ms.mu.Lock()
	current := ms.gauge[name] + value
	ms.gauge[name] = current
	ms.mu.Unlock()
	log.Printf(
		"Add gauge metric: name=%q value=%v currentValue=%v\n",
		name, value, current,
	)
}
