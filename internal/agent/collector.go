package agent

import (
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/field"
)

// `t` is metric type "gauge" or "counter", `v` is metric value
type metric struct {
	t, v string
}

// return point on metric `t` is one of "gauge" or "counter"
func newMetric(t string) *metric {
	return &metric{t: t}
}

var memMetrics = map[string]*metric{
	"Alloc":         newMetric(server.Gauge),
	"BuckHashSys":   newMetric(server.Gauge),
	"Frees":         newMetric(server.Gauge),
	"GCCPUFraction": newMetric(server.Gauge),
	"GCSys":         newMetric(server.Gauge),
	"HeapAlloc":     newMetric(server.Gauge),
	"HeapIdle":      newMetric(server.Gauge),
	"HeapInuse":     newMetric(server.Gauge),
	"HeapObjects":   newMetric(server.Gauge),
	"HeapReleased":  newMetric(server.Gauge),
	"HeapSys":       newMetric(server.Gauge),
	"LastGC":        newMetric(server.Gauge),
	"Lookups":       newMetric(server.Gauge),
	"MCacheInuse":   newMetric(server.Gauge),
	"MCacheSys":     newMetric(server.Gauge),
	"MSpanInuse":    newMetric(server.Gauge),
	"MSpanSys":      newMetric(server.Gauge),
	"Mallocs":       newMetric(server.Gauge),
	"NextGC":        newMetric(server.Gauge),
	"NumForcedGC":   newMetric(server.Gauge),
	"NumGC":         newMetric(server.Gauge),
	"OtherSys":      newMetric(server.Gauge),
	"PauseTotalNs":  newMetric(server.Gauge),
	"StackInuse":    newMetric(server.Gauge),
	"StackSys":      newMetric(server.Gauge),
	"Sys":           newMetric(server.Gauge),
	"TotalAlloc":    newMetric(server.Gauge),
}

const (
	RandomValue = "RandomValue"
	PollCount   = "PollCount"
)

var extraMetrics = map[string]*metric{
	RandomValue: newMetric(server.Gauge),
	PollCount:   newMetric(server.Counter),
}

var next = newCounter(0)

type Collector struct {
	pollInt, reportInt time.Duration
	handler            func()
}

func NewCollector(pollInt, reropInt time.Duration, handler func()) *Collector {
	return &Collector{pollInt, reropInt, handler}
}

func (c *Collector) Run() {
	log.Printf(
		"Run collector with intervals: poll = %vs, report = %vs\n",
		c.pollInt.Seconds(), c.reportInt.Seconds(),
	)

	var untilReport = time.Duration(0)

	for {
		getMemMetrics(memMetrics)
		getExtraMetrics(extraMetrics)

		if untilReport.Seconds() == 0 {
			c.handler()
			untilReport = time.Duration(c.reportInt)
		}
		time.Sleep(c.pollInt)
		untilReport -= time.Duration(2 * time.Second)
	}
}

func getMemMetrics(memMetrics map[string]*metric) {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)

	for name, metric := range memMetrics {
		v, err := field.Value(memStat, name)
		if err != nil {
			log.Println(err)
		}

		switch v := v.(type) {
		case float64:
			metric.v = strconv.FormatFloat(v, 'f', -1, 64)
		case uint64:
			metric.v = strconv.FormatFloat(float64(v), 'f', -1, 64)
		case uint32:
			metric.v = strconv.FormatFloat(float64(v), 'f', -1, 64)
		default:
			log.Printf("Not converted: name=%s, type=%T, value=%v\n", name, v, v)
		}
	}
}

func getExtraMetrics(map[string]*metric) {
	extraMetrics[RandomValue].v = strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
	extraMetrics[PollCount].v = strconv.Itoa(next())
}

func newCounter(n int) func() int {
	return func() int {
		n++
		return n
	}
}
