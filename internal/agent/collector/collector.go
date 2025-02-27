package collector

type metricsData struct {
	counter map[string]int64
	gauge   map[string]float64
}

func newMetricsData() metricsData {
	return metricsData{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}
