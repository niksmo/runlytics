package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/metrics"
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

func newMemory(
	file *os.File, interval time.Duration, restore bool,
) *memoryStorage {
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

func (ms *memoryStorage) CheckDB(_ context.Context) error {
	return errors.New("database: is not used")
}

func (ms *memoryStorage) UpdateCounterByName(
	_ context.Context, name string, value int64,
) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	prev := ms.data.Counter[name]
	current := prev + value
	ms.data.Counter[name] = current
	return current, nil
}

func (ms *memoryStorage) UpdateGaugeByName(
	_ context.Context, name string, value float64,
) (float64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data.Gauge[name] = value
	return value, nil
}

func (ms *memoryStorage) UpdateCounterList(ctx context.Context, mSlice []metrics.MetricsCounter) error {
	for _, item := range mSlice {
		ms.UpdateCounterByName(ctx, item.ID, item.Delta)
	}
	return nil
}

func (ms *memoryStorage) UpdateGaugeList(ctx context.Context, mSlice []metrics.MetricsGauge) error {
	for _, item := range mSlice {
		ms.UpdateGaugeByName(ctx, item.ID, item.Value)
	}
	return nil
}

func (ms *memoryStorage) ReadCounterByName(
	_ context.Context, name string,
) (int64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.data.Counter[name]

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
	}
	return value, nil
}

func (ms *memoryStorage) ReadGaugeByName(
	_ context.Context, name string,
) (float64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.data.Gauge[name]

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
	}
	return value, nil
}

func (ms *memoryStorage) ReadGauge(
	_ context.Context,
) (map[string]float64, error) {
	gauge := make(map[string]float64, len(ms.data.Gauge))

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for k, v := range ms.data.Gauge {
		gauge[k] = v
	}

	return gauge, nil
}

func (ms *memoryStorage) ReadCounter(
	_ context.Context,
) (map[string]int64, error) {
	counter := make(map[string]int64, len(ms.data.Counter))

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for k, v := range ms.data.Counter {
		counter[k] = v
	}

	return counter, nil
}

// Restoring file, starting save interval and waiting graceful shutdown
func (ms *memoryStorage) Run(ctx context.Context, wg *sync.WaitGroup) {
	ms.restoreFile()

	if !ms.isSync() {
		go ms.intervalSave()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		ms.close()
	}()
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
	} else {
		logger.Log.Debug("Storage file closed properly")
	}
}
