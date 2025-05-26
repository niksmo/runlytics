// Package di provides interfaces for project dependency injection.
package di

import (
	"bytes"
	"context"
	"sync"

	"github.com/niksmo/runlytics/pkg/metrics"
)

// GaugeMetricsGetter is the interface that wraps the GetGaugeMetrics method.
type GaugeMetricsGetter interface {
	GetGaugeMetrics() map[string]float64
}

// CounterMetricsGetter is the interface
// that wraps the GetCounterMetrics method.
type CounterMetricsGetter interface {
	GetCounterMetrics() map[string]int64
}

// Runner is the interface that wraps the Run method.
type Runner interface {
	Run()
}

// MetricsCollector is the interface that groups
// the GetGaugeMetrics, GetCounterMetrics and Run methods.
type MetricsCollector interface {
	Runner
	GaugeMetricsGetter
	CounterMetricsGetter
}

// Looger is the interface that wraps the Debugw, Infow and Errorw methods.
type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

// FileCloser is the interface that wraps the basic Close method.
type FileCloser interface {
	Close() error
}

// FileClearer is the interface that wraps the basic Clear method.
type FileClearer interface {
	Clear() error
}

// FileLoader is the interface that wraps the basic Load method.
type FileLoader interface {
	Load() ([]byte, error)
}

// FileSaver is the interface that wraps the basic Save method.
type FileSaver interface {
	Save([]byte) error
}

// FileOperator is the interface that wraps the
// Clear, Load, Save and Close methods.
type FileOperator interface {
	FileClearer
	FileCloser
	FileLoader
	FileSaver
}

// UpdateByNameRepository is the interface that wraps the
// UpdateCounterByName and UpdateGaugeByName methods.
type UpdateByNameRepository interface {
	UpdateCounterByName(ctx context.Context, name string, value int64) (int64, error)
	UpdateGaugeByName(ctx context.Context, name string, value float64) (float64, error)
}

// BatchUpdate is the interface that wraps the
// UpdateCounterList and UpdateGaugeList methods.
type BatchUpdate interface {
	UpdateCounterList(ctx context.Context, slice metrics.MetricsList) error
	UpdateGaugeList(ctx context.Context, slice metrics.MetricsList) error
}

// ReadByNameRepository is the interface that wraps the
// ReadCounterByName and ReadGaugeByName methods.
type ReadByNameRepository interface {
	ReadCounterByName(ctx context.Context, name string) (int64, error)
	ReadGaugeByName(ctx context.Context, name string) (float64, error)
}

// ReadListRepository is the interface that wraps the
// ReadGauge and ReadCounter methods.
type ReadListRepository interface {
	ReadGauge(context.Context) (map[string]float64, error)
	ReadCounter(context.Context) (map[string]int64, error)
}

// Repository is the interface that groups
// the UpdateCounterByName, UpdateGaugeByName, UpdateCounterList,
// UpdateGaugeList,ReadCounterByName, ReadGaugeByName,
// ReadGauge, ReadCounter and Run methods.
type Repository interface {
	ReadByNameRepository
	ReadListRepository
	UpdateByNameRepository
	BatchUpdate
	Run(stopCtx context.Context, wg *sync.WaitGroup)
}

// HealthCheckService is the interface that wraps the Check method.
type HealthCheckService interface {
	Check(ctx context.Context) error
}

// HTMLService is the interface that wraps the RenderMetricsList method.
type HTMLService interface {
	RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error
}

// ReadService is the interface that wraps the Read method.
type ReadService interface {
	Read(context.Context, *metrics.Metrics) error
}

// UpdateService is the interface that wraps the Update method.
type UpdateService interface {
	Update(context.Context, *metrics.Metrics) error
}

// BatchUpdateService is the interface that wraps the BatchUpdate method.
type BatchUpdateService interface {
	BatchUpdate(context.Context, metrics.MetricsList) error
}

// IID is the interface that wraps the ID method.
type IID interface {
	ID() int64
}

// IPayload is the interface that wraps the Payload method.
type IPayload interface {
	Payload() []metrics.Metrics
}

// IErr is the interface that wraps the Err method.
type IErr interface {
	Err() error
}

// Job is the interface that wraps the ID and Payload methods.
type Job interface {
	IID
	IPayload
}

// JobErr is the interface that wraps the ID and Err methods.
type JobErr interface {
	IID
	IErr
}

// Decrypter is the interface that wraps the DecryptMsg method.
type Decrypter interface {
	DecryptMsg([]byte) ([]byte, error)
}

// Encrypter is the interface that wraps the EncryptMsg method.
type Encrypter interface {
	EncryptMsg([]byte) ([]byte, error)
}
