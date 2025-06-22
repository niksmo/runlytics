// Package di provides interfaces for project dependency injection.
package di

import (
	"bytes"
	"context"

	"github.com/niksmo/runlytics/pkg/metrics"
)

// MetricsGetter is the interface that wraps the GetGaugeMetrics method.
type MetricsGetter interface {
	GetMetrics() metrics.MetricsList
}

type SendMetricsFunc func(ctx context.Context, m metrics.MetricsList, enc Encrypter, url, key, ip string) error

type Runner interface {
	Run()
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

type RunStopper interface {
	Runner
	Stopper
}

type MustRunStopper interface {
	MustRunner
	Stopper
}

// MetricsProvider is the interface that groups
// the GetMetrics, Run and Stop methods.
type MetricsProvider interface {
	MetricsGetter
	Stopper
	Runner
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
	MustRunner
	Pinger
	Stopper
}

// IHealthCheckService is the interface that wraps the Check method.
type IHealthCheckService interface {
	Check(ctx context.Context) error
}

// IHTMLService is the interface that wraps the RenderMetricsList method.
type IHTMLService interface {
	RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error
}

// IReadService is the interface that wraps the Read method.
type IReadService interface {
	Read(context.Context, *metrics.Metrics) error
}

// IUpdateService is the interface that wraps the Update method.
type IUpdateService interface {
	Update(context.Context, *metrics.Metrics) error
}

// IBatchUpdateService is the interface that wraps the BatchUpdate method.
type IBatchUpdateService interface {
	BatchUpdate(context.Context, metrics.MetricsList) error
}

// Decrypter is the interface that wraps the DecryptMsg method.
type Decrypter interface {
	DecryptMsg([]byte) ([]byte, error)
}

// Encrypter is the interface that wraps the EncryptMsg method.
type Encrypter interface {
	EncryptMsg([]byte) ([]byte, error)
}
