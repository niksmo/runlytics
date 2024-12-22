package repository

import "log"

type memStorage struct {
	count map[string]int64
	gauge map[string]float64
}

func NewMemStorage() *memStorage {
	return &memStorage{
		make(map[string]int64),
		make(map[string]float64),
	}
}

func (ms *memStorage) AddCounter(name string, value int64) {
	ms.count[name] += value
	log.Printf(
		"Add count metric: name=%q value=%v currentValue=%v\n",
		name, value, ms.count[name],
	)
}

func (ms *memStorage) AddGauge(name string, value float64) {
	ms.gauge[name] = value
	log.Printf(
		"Add gauge metric: name=%q value=%v currentValue=%v\n",
		name, value, ms.gauge[name],
	)
}
