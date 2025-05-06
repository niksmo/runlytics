package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"maps"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type storageData struct {
	Counter map[string]int64   `json:"counter"`
	Gauge   map[string]float64 `json:"gauge"`
}

// MemoryStorage store metrics in underlyin map and implements [di.Repository] interface.
type MemoryStorage struct {
	mu       sync.RWMutex
	fo       di.FileOperator
	data     storageData
	restore  bool
	interval time.Duration
}

// NewMemory returns MemoryStorage pointer.
func NewMemory(
	fo di.FileOperator, interval time.Duration, restore bool,
) *MemoryStorage {
	ms := MemoryStorage{
		data: storageData{
			Counter: make(map[string]int64),
			Gauge:   make(map[string]float64),
		},
		interval: interval,
		fo:       fo,
		restore:  restore,
	}
	return &ms
}

// Run starts [MemoryStorage] and then waiting graceful shutdown.
//
// If MemoryStorage.restore is true, restores metrics data from file.
func (ms *MemoryStorage) Run(stopCtx context.Context, wg *sync.WaitGroup) {
	ms.restoreData()

	if !ms.isSync() {
		go ms.intervalSave()
	}

	wg.Add(1)
	go ms.waitStop(stopCtx, wg)
}

// UpdateCounterByName returns updated counter value and nil error.
func (ms *MemoryStorage) UpdateCounterByName(
	_ context.Context, name string, value int64,
) (int64, error) {
	ms.mu.Lock()
	prev := ms.data.Counter[name]
	current := prev + value
	ms.data.Counter[name] = current
	ms.mu.Unlock()

	if ms.isSync() {
		ms.save()
	}
	return current, nil
}

// UpdateGaugeByName returns updated gauge value and nil error.
func (ms *MemoryStorage) UpdateGaugeByName(
	_ context.Context, name string, value float64,
) (float64, error) {
	ms.mu.Lock()
	ms.data.Gauge[name] = value
	ms.mu.Unlock()

	if ms.isSync() {
		ms.save()
	}
	return value, nil
}

// UpdateCounterList returns nil error.
func (ms *MemoryStorage) UpdateCounterList(
	ctx context.Context, mSlice metrics.MetricsList,
) error {
	for _, item := range mSlice {
		ms.UpdateCounterByName(ctx, item.ID, *item.Delta)
	}
	return nil
}

// UpdateGaugeList returns nil error.
func (ms *MemoryStorage) UpdateGaugeList(
	ctx context.Context, mSlice metrics.MetricsList,
) error {
	for _, item := range mSlice {
		ms.UpdateGaugeByName(ctx, item.ID, *item.Value)
	}
	return nil
}

// ReadCounterByName returns counter value and nil error.
func (ms *MemoryStorage) ReadCounterByName(
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

// ReadGaugeByName returns gauge value and nil error
func (ms *MemoryStorage) ReadGaugeByName(
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

// ReadGauge returns gauge metrics copy.
func (ms *MemoryStorage) ReadGauge(
	_ context.Context,
) (map[string]float64, error) {
	gauge := make(map[string]float64, len(ms.data.Gauge))

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	maps.Copy(gauge, ms.data.Gauge)

	return gauge, nil
}

// ReadCounter returns counter metrics copy.
func (ms *MemoryStorage) ReadCounter(
	_ context.Context,
) (map[string]int64, error) {
	counter := make(map[string]int64, len(ms.data.Counter))

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	maps.Copy(counter, ms.data.Counter)

	return counter, nil
}

func (ms *MemoryStorage) restoreData() {

	if !ms.restore {
		ms.fo.Clear()
		return
	}

	rawData, err := ms.fo.Load()
	if err != nil {
		logger.Log.Error("FileOperator load", zap.Error(err))
		return
	}
	if len(rawData) == 0 {
		logger.Log.Info("Storage file is empty")
		return
	}

	var data storageData
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		logger.Log.Error("JSON unmarshal", zap.Error(err))
	}

	ms.mu.Lock()
	ms.data = data
	ms.mu.Unlock()
	nMetrics := len(data.Counter) + len(data.Gauge)

	logger.Log.Info("Restore metrics", zap.Int("count", nMetrics))
}

func (ms *MemoryStorage) isSync() bool {
	return ms.interval == 0
}

func (ms *MemoryStorage) save() {
	ms.mu.RLock()
	rawData, err := json.Marshal(ms.data)
	if err != nil {
		logger.Log.Error("JSON marshal", zap.Error(err))
	}
	if err = ms.fo.Save(rawData); err != nil {
		logger.Log.Error("FileOperator save", zap.Error(err))
	}
	ms.mu.RUnlock()
	logger.Log.Debug("Save metrics to file")
}

func (ms *MemoryStorage) intervalSave() {
	for {
		time.Sleep(ms.interval)
		ms.save()
	}
}

func (ms *MemoryStorage) close() {
	ms.save()
	if err := ms.fo.Close(); err != nil {
		logger.Log.Error("FileOperator close", zap.Error(err))
	} else {
		logger.Log.Debug("FileOperator closed properly")
	}
}

func (ms *MemoryStorage) waitStop(
	stopCtx context.Context, wg *sync.WaitGroup,
) {
	defer wg.Done()
	<-stopCtx.Done()
	ms.close()
}
