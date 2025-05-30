// Package di provides interfaces for project dependency injection.
package di

import (
	"bytes"
	"context"

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
	Run() error
}

type MustRunner interface {
	MustRun()
}

type Pinger interface {
	Ping(context.Context) error
}

type Stopper interface {
	Stop()
}

type MustRunStopper interface {
	MustRunner
	Stopper
}

// MetricsCollector is the interface that groups
// the GetGaugeMetrics, GetCounterMetrics and Run methods.
type IMetricsCollector interface {
	Runner
	GaugeMetricsGetter
	CounterMetricsGetter
}

// Closer is the interface that wraps the basic Close method.
type Closer interface {
	Close() error
}

// IClear is the interface that wraps the basic Clear method.
type IClear interface {
	Clear() error
}

// Loader is the interface that wraps the basic Load method.
type Loader interface {
	Load() ([]byte, error)
}

// Saver is the interface that wraps the basic Save method.
type Saver interface {
	Save([]byte) error
}

// FileOperator is the interface that wraps the
// Clear, Load, Save and Close methods.
type FileOperator interface {
	IClear
	Closer
	Loader
	Saver
}

// IUpdateByNameStorage is the interface that wraps the
// UpdateCounterByName and UpdateGaugeByName methods.
type IUpdateByNameStorage interface {
	UpdateCounterByName(ctx context.Context, name string, value int64) (int64, error)
	UpdateGaugeByName(ctx context.Context, name string, value float64) (float64, error)
}

// IBatchUpdateStorage is the interface that wraps the
// UpdateCounterList and UpdateGaugeList methods.
type IBatchUpdateStorage interface {
	UpdateCounterList(ctx context.Context, slice metrics.MetricsList) error
	UpdateGaugeList(ctx context.Context, slice metrics.MetricsList) error
}

// IReadByNameStorage is the interface that wraps the
// ReadCounterByName and ReadGaugeByName methods.
type IReadByNameStorage interface {
	ReadCounterByName(ctx context.Context, name string) (int64, error)
	ReadGaugeByName(ctx context.Context, name string) (float64, error)
}

// IReadListStorage is the interface that wraps the
// ReadGauge and ReadCounter methods.
type IReadListStorage interface {
	ReadGauge(context.Context) (map[string]float64, error)
	ReadCounter(context.Context) (map[string]int64, error)
}

// Storage is the interface that groups
// the UpdateCounterByName, UpdateGaugeByName, UpdateCounterList,
// UpdateGaugeList,ReadCounterByName, ReadGaugeByName,
// ReadGauge, ReadCounter and Run methods.
type IStorage interface {
	IReadByNameStorage
	IReadListStorage
	IUpdateByNameStorage
	IBatchUpdateStorage
	Runner
	MustRunner
	Pinger
	Stopper
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
type IJob interface {
	IID
	IPayload
}

// JobErr is the interface that wraps the ID and Err methods.
type IJobErr interface {
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
