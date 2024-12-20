package storage

import "log"

type MemStorage struct {
	count map[string]int
	gauge map[string]float64
}

func (ms *MemStorage) addCount(name string, value int) {
	ms.count[name] += value
	log.Printf(
		"Add count metric: name=%q value=%v currentValue=%v\n",
		name, value, ms.count[name],
	)
}

func (ms *MemStorage) addGauge(name string, value float64) {
	ms.gauge[name] = value
	log.Printf(
		"Add gauge metric: name=%q value=%v currentValue=%v\n",
		name, value, ms.gauge[name],
	)
}
