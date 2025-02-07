package di

import (
	"bytes"
	"context"
	"os"
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

type GaugeCounterMetricsGetter interface {
	GaugeMetricsGetter
	CounterMetricsGetter
}

type Logger interface {
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

type Config interface {
	IsDatabase() bool
	File() *os.File
	SaveInterval() time.Duration
	Restore() bool
}

type Verifier interface {
	Verify() error
}

type SchemeVerifier interface {
	VerifyScheme(Verifier) error
}

type MetricsParamsVerifier interface {
	VerifyParams(id, mType, value string) (*metrics.MetricsUpdate, error)
}

type MetricsParamsSchemeVerifier interface {
	MetricsParamsVerifier
	SchemeVerifier
}

type UpdateRepository interface {
	UpdateCounterByName(ctx context.Context, name string, value int64) (int64, error)
	UpdateGaugeByName(ctx context.Context, name string, value float64) (float64, error)
}

type ReadByNameRepository interface {
	ReadCounterByName(ctx context.Context, name string) (int64, error)
	ReadGaugeByName(ctx context.Context, name string) (float64, error)
}

type ReadRepository interface {
	ReadGauge(context.Context) (map[string]float64, error)
	ReadCounter(context.Context) (map[string]int64, error)
}

type Repository interface {
	UpdateRepository
	ReadByNameRepository
	ReadRepository
	Run(stopCtx context.Context, wg *sync.WaitGroup)
}

type HealthCheckService interface {
	Check(ctx context.Context) error
}

type HTMLService interface {
	RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error
}

type UpdateService interface {
	Update(context.Context, *metrics.MetricsUpdate) (metrics.Metrics, error)
}

type ReadService interface {
	Read(ctx context.Context, mData *metrics.MetricsRead) (metrics.Metrics, error)
}
