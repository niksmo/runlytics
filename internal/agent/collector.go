package agent

import (
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/niksmo/runlytics/pkg/field"
)

const (
	gauge   = "gauge"
	counter = "counter"
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
	"Alloc":         newMetric(gauge),
	"BuckHashSys":   newMetric(gauge),
	"Frees":         newMetric(gauge),
	"GCCPUFraction": newMetric(gauge),
	"GCSys":         newMetric(gauge),
	"HeapAlloc":     newMetric(gauge),
	"HeapIdle":      newMetric(gauge),
	"HeapInuse":     newMetric(gauge),
	"HeapObjects":   newMetric(gauge),
	"HeapReleased":  newMetric(gauge),
	"HeapSys":       newMetric(gauge),
	"LastGC":        newMetric(gauge),
	"Lookups":       newMetric(gauge),
	"MCacheInuse":   newMetric(gauge),
	"MCacheSys":     newMetric(gauge),
	"MSpanInuse":    newMetric(gauge),
	"MSpanSys":      newMetric(gauge),
	"Mallocs":       newMetric(gauge),
	"NextGC":        newMetric(gauge),
	"NumForcedGC":   newMetric(gauge),
	"NumGC":         newMetric(gauge),
	"OtherSys":      newMetric(gauge),
	"PauseTotalNs":  newMetric(gauge),
	"StackInuse":    newMetric(gauge),
	"StackSys":      newMetric(gauge),
	"Sys":           newMetric(gauge),
	"TotalAlloc":    newMetric(gauge),
}

const (
	RandomValue = "RandomValue"
	PollCount   = "PollCount"
)

var extraMetrics = map[string]*metric{
	RandomValue: newMetric(gauge),
	PollCount:   newMetric(counter),
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
