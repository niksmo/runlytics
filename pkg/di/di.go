package di

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/niksmo/runlytics/pkg/metrics"
)

type GaugeMetricsGetter interface {
	GetGaugeMetrics() map[string]float64
}

type CounterMetricsGetter interface {
	GetCounterMetrics() map[string]int64
}

type Runner interface {
	Run()
}

type MetricsCollector interface {
	Runner
	GaugeMetricsGetter
	CounterMetricsGetter
}

type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

type ServerConfig interface {
	IsDatabase() bool
	SaveInterval() time.Duration
	Restore() bool
}

type FileOperator interface {
	Clear() (err error)
	Load() ([]byte, error)
	Save([]byte) (err error)
	Close() error
}

type UpdateByNameRepository interface {
	UpdateCounterByName(ctx context.Context, name string, value int64) (int64, error)
	UpdateGaugeByName(ctx context.Context, name string, value float64) (float64, error)
}

type BatchUpdate interface {
	UpdateCounterList(ctx context.Context, slice []metrics.Metrics) error
	UpdateGaugeList(ctx context.Context, slice []metrics.Metrics) error
}

type ReadByNameRepository interface {
	ReadCounterByName(ctx context.Context, name string) (int64, error)
	ReadGaugeByName(ctx context.Context, name string) (float64, error)
}

type ReadListRepository interface {
	ReadGauge(context.Context) (map[string]float64, error)
	ReadCounter(context.Context) (map[string]int64, error)
}

type Repository interface {
	ReadByNameRepository
	ReadListRepository
	UpdateByNameRepository
	BatchUpdate
	Run(stopCtx context.Context, wg *sync.WaitGroup)
}

type HealthCheckService interface {
	Check(ctx context.Context) error
}

type HTMLService interface {
	RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error
}

type ReadService interface {
	Read(context.Context, *metrics.Metrics) error
}

type UpdateService interface {
	Update(context.Context, *metrics.Metrics) error
}

type BatchUpdateService interface {
	BatchUpdate(context.Context, metrics.MetricsList) error
}

type Job interface {
	ID() int64
	Payload() []metrics.Metrics
}

type JobErr interface {
	ID() int64
	Err() error
}
