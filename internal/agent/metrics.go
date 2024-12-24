package agent

import (
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/field"
)

const (
	RandomValue = "RandomValue"
	PollCount   = "PollCount"
)

type Metrics struct {
	mu   sync.Mutex
	data map[string]Metric
}

// `T` is one of metric type "gauge" or "counter", `V` is metric value
type Metric struct {
	T, V string
}

var memMetrics = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

var next = newCounter(0)

func getMemMetrics(metrics map[string]Metric) {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)

	for _, name := range memMetrics {
		v, err := field.Value(memStat, name)
		if err != nil {
			log.Println(err)
		}

		var statValue string

		switch v := v.(type) {
		case float64:
			statValue = strconv.FormatFloat(v, 'f', -1, 64)
		case uint64:
			statValue = strconv.FormatFloat(float64(v), 'f', -1, 64)
		case uint32:
			statValue = strconv.FormatFloat(float64(v), 'f', -1, 64)
		default:
			log.Printf("Not converted: name=%s, type=%T, value=%v\n", name, v, v)
		}

		if statValue != "" {
			metrics[name] = Metric{server.Gauge, statValue}
		}
	}
}

func getExtraMetrics(metrics map[string]Metric) {
	metrics[RandomValue] = Metric{server.Gauge, strconv.FormatFloat(rand.Float64(), 'f', -1, 64)}
	metrics[PollCount] = Metric{server.Counter, strconv.Itoa(next())}
}

func newCounter(n int) func() int {
	return func() int {
		n++
		return n
	}
}
