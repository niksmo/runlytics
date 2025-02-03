package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type storageData struct {
	Counter map[string]int64   `json:"counter"`
	Gauge   map[string]float64 `json:"gauge"`
}

type memoryStorage struct {
	mu       sync.RWMutex
	data     storageData
	interval time.Duration
	file     *os.File
	restore  bool
	encoder  *json.Encoder
	decoder  *json.Decoder
}

func NewMemory(file *os.File, interval time.Duration, restore bool) *memoryStorage {
	ms := memoryStorage{
		data: storageData{
			Counter: make(map[string]int64),
			Gauge:   make(map[string]float64),
		},
		interval: interval,
		file:     file,
		restore:  restore,
		encoder:  json.NewEncoder(file),
	}

	if restore {
		ms.decoder = json.NewDecoder(file)
	}
	return &ms
}

func (ms *memoryStorage) UpdateCounterByName(name string, value int64) (int64, error) {
	ms.mu.Lock()
	prev := ms.data.Counter[name]
	current := prev + value
	ms.data.Counter[name] = current
	ms.mu.Unlock()
	return current, nil
}

func (ms *memoryStorage) UpdateGaugeByName(name string, value float64) (float64, error) {
	ms.mu.Lock()
	ms.data.Gauge[name] = value
	ms.mu.Unlock()
	return value, nil
}

func (ms *memoryStorage) ReadCounterByName(name string) (int64, error) {
	ms.mu.RLock()
	value, ok := ms.data.Counter[name]
	ms.mu.RUnlock()

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	}
	return value, nil
}

func (ms *memoryStorage) ReadGaugeByName(name string) (float64, error) {
	ms.mu.RLock()
	value, ok := ms.data.Gauge[name]
	ms.mu.RUnlock()

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, ErrNotExists)
	}
	return value, nil
}

func (ms *memoryStorage) ReadGauge() (map[string]float64, error) {
	gauge := make(map[string]float64, len(ms.data.Gauge))

	ms.mu.RLock()
	for k, v := range ms.data.Gauge {
		gauge[k] = v
	}
	ms.mu.RUnlock()

	return gauge, nil
}

func (ms *memoryStorage) ReadCounter() (map[string]int64, error) {
	counter := make(map[string]int64, len(ms.data.Counter))

	ms.mu.RLock()
	for k, v := range ms.data.Counter {
		counter[k] = v
	}
	ms.mu.RUnlock()

	return counter, nil
}

func (ms *memoryStorage) Run() {
	ms.restoreFile()

	if !ms.isSync() {
		go ms.intervalSave()
	}

	go ms.interceptSigInt()
}

func (ms *memoryStorage) restoreFile() {

	if !ms.restore {
		ms.file.Truncate(0)
		return
	}

	var data storageData
	err := ms.decoder.Decode(&data)
	if errors.Is(err, io.EOF) {
		logger.Log.Info("File storage is empty")
		return
	}

	if err != nil {
		logger.Log.Error("JSON decode", zap.Error(err))
		return
	}

	ms.mu.Lock()
	ms.data = data
	ms.mu.Unlock()
	nMetrics := len(data.Counter) + len(data.Gauge)

	logger.Log.Info(
		"Restore metrics",
		zap.Int("count", nMetrics),
	)
}

func (ms *memoryStorage) isSync() bool {
	return ms.interval == 0
}

func (ms *memoryStorage) save() {
	ms.file.Seek(0, 0)
	ms.mu.RLock()
	if err := ms.encoder.Encode(ms.data); err != nil {
		logger.Log.Error("JSON encode", zap.Error(err))
	}
	ms.mu.RUnlock()
	logger.Log.Debug("Save metrics to file")
}

func (ms *memoryStorage) intervalSave() {
	for {
		time.Sleep(ms.interval)
		ms.save()
	}
}

func (ms *memoryStorage) close() {
	ms.save()
	if err := ms.file.Close(); err != nil {
		logger.Log.Error("Close storage file", zap.Error(err))
	}
}

func (ms *memoryStorage) interceptSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	sig := <-c
	logger.Log.Debug("Got signal", zap.String("signal", sig.String()))
	ms.close()
	os.Exit(0)
}
