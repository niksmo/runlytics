package filestorage

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

type data struct {
	Counter map[string]int64   `json:"counter"`
	Gauge   map[string]float64 `json:"gauge"`
}

// FileStorage store metrics in underlyin map and implements [di.Storage] interface.
type FileStorage struct {
	mu       sync.RWMutex
	fo       di.FileOperator
	data     data
	restore  bool
	interval time.Duration
	ticker   *time.Ticker
}

// New returns MemoryStorage pointer.
func New(
	fo di.FileOperator, interval time.Duration, restore bool,
) *FileStorage {
	return &FileStorage{
		data: data{
			Counter: make(map[string]int64),
			Gauge:   make(map[string]float64),
		},
		interval: interval,
		fo:       fo,
		restore:  restore,
	}
}

func (fs *FileStorage) MustRun() {
	if err := fs.Run(); err != nil {
		panic(err)
	}
}

// Run starts [FileStorage] and then waiting graceful shutdown.
//
// If MemoryStorage.restore is true, restores metrics data from file.
func (fs *FileStorage) Run() error {
	const op = "filestorage.Run"
	var err error
	if err = fs.restoreData(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !fs.isSync() {
		fs.ticker = time.NewTicker(fs.interval)
		go fs.intervalSave(func(saveErr error) {
			err = saveErr
		})
	}

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (fs *FileStorage) Ping(context.Context) error {
	return nil
}

func (fs *FileStorage) Stop() {
	const op = "filestorage.Stop"
	log := logger.Log.With(zap.String("op", op))

	log.Info("filestorage stopping gracefully")

	if !fs.isSync() {
		fs.ticker.Stop()
	}

	if err := fs.save(); err != nil {
		log.Error("failed to save", zap.Error(err))
		return
	}

	if err := fs.fo.Close(); err != nil {
		log.Error("failed to close fileoperator", zap.Error(err))
		return
	}
	log.Info("filestorage stopped")
}

// UpdateCounterByName returns updated counter value and nil error.
func (fs *FileStorage) UpdateCounterByName(
	_ context.Context, name string, value int64,
) (int64, error) {
	fs.mu.Lock()
	prev := fs.data.Counter[name]
	current := prev + value
	fs.data.Counter[name] = current
	fs.mu.Unlock()

	if fs.isSync() {
		fs.save()
	}
	return current, nil
}

// UpdateGaugeByName returns updated gauge value and nil error.
func (fs *FileStorage) UpdateGaugeByName(
	_ context.Context, name string, value float64,
) (float64, error) {
	fs.mu.Lock()
	fs.data.Gauge[name] = value
	fs.mu.Unlock()

	if fs.isSync() {
		fs.save()
	}
	return value, nil
}

// UpdateCounterList returns nil error.
func (fs *FileStorage) UpdateCounterList(
	ctx context.Context, mSlice metrics.MetricsList,
) error {
	for _, item := range mSlice {
		fs.UpdateCounterByName(ctx, item.ID, *item.Delta)
	}
	return nil
}

// UpdateGaugeList returns nil error.
func (fs *FileStorage) UpdateGaugeList(
	ctx context.Context, mSlice metrics.MetricsList,
) error {
	for _, item := range mSlice {
		fs.UpdateGaugeByName(ctx, item.ID, *item.Value)
	}
	return nil
}

// ReadCounterByName returns counter value and nil error.
func (fs *FileStorage) ReadCounterByName(
	_ context.Context, name string,
) (int64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	value, ok := fs.data.Counter[name]

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
	}
	return value, nil
}

// ReadGaugeByName returns gauge value and nil error
func (fs *FileStorage) ReadGaugeByName(
	_ context.Context, name string,
) (float64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	value, ok := fs.data.Gauge[name]

	if !ok {
		return 0, fmt.Errorf("metric '%s' is %w", name, server.ErrNotExists)
	}
	return value, nil
}

// ReadGauge returns gauge metrics copy.
func (fs *FileStorage) ReadGauge(
	_ context.Context,
) (map[string]float64, error) {
	gauge := make(map[string]float64, len(fs.data.Gauge))

	fs.mu.RLock()
	defer fs.mu.RUnlock()
	maps.Copy(gauge, fs.data.Gauge)

	return gauge, nil
}

// ReadCounter returns counter metrics copy.
func (fs *FileStorage) ReadCounter(
	_ context.Context,
) (map[string]int64, error) {
	counter := make(map[string]int64, len(fs.data.Counter))

	fs.mu.RLock()
	defer fs.mu.RUnlock()
	maps.Copy(counter, fs.data.Counter)

	return counter, nil
}

func (fs *FileStorage) restoreData() error {
	if !fs.restore {
		if err := fs.fo.Clear(); err != nil {
			return err
		}
	}

	loaded, err := fs.fo.Load()
	if err != nil {
		return err
	}
	if len(loaded) == 0 {
		return nil
	}

	var data data
	err = json.Unmarshal(loaded, &data)
	if err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.data = data
	return nil
}

func (fs *FileStorage) isSync() bool {
	return fs.interval == 0
}

func (fs *FileStorage) save() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	data, err := json.Marshal(fs.data)
	if err != nil {
		return err
	}
	if err = fs.fo.Save(data); err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) intervalSave(errFn func(error)) {
	for range fs.ticker.C {
		if err := fs.save(); err != nil {
			errFn(err)
			return
		}
	}
}
