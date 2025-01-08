package server

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

type RepoUpdate interface {
	AddCounter(name string, value int64)
	SetGauge(name string, value float64)
}

type RepoReadByName interface {
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
}

type RepoRead interface {
	GetData() (map[string]float64, map[string]int64)
}
