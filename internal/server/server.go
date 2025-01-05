package server

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

type Repository interface {
	AddCounter(name string, value int64)
	AddGauge(name string, value float64)
}
