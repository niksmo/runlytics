package server

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

type RepositoryUpdate interface {
	AddCounter(name string, value int64)
	SetGauge(name string, value float64)
}

type RepositoryRead interface {
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
}
