package agent

import (
	"errors"
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/counter"
	"github.com/niksmo/runlytics/pkg/field"
)

var count = counter.New(0)

var (
	ErrReportLessPoll   = errors.New("report interval should be more or equal to poll interval")
	ErrMinIntervalValue = errors.New("both intervals should be more or equal 1s")
)

type ReportHandler func(data []Metric)

type collector struct {
	poll, report time.Duration
	rh           ReportHandler
	data         []Metric
}

func NewCollector(poll, report time.Duration, rh ReportHandler) (*collector, error) {
	s := time.Duration(1 * time.Second)
	if poll < s || report < s {
		return nil, ErrMinIntervalValue
	}

	if report < poll {
		return nil, ErrReportLessPoll
	}

	return &collector{poll: poll, report: report, rh: rh}, nil
}

func (c *collector) getData() []Metric {
	ret := make([]Metric, len(c.data))
	copy(ret, c.data)
	return ret
}

func (c *collector) collectMetrics() {
	memMetrics := getMemMetrics()
	extraMetrics := getExtraMetrics()
	c.data = append(memMetrics, extraMetrics...)
}

func (c *collector) Run() {
	log.Printf(
		"Run metrics collector with intervals: poll = %vs, report = %vs\n",
		c.poll.Seconds(), c.report.Seconds(),
	)

	go func() {
		for {
			c.collectMetrics()
			log.Println("[POLL]Wait for", c.poll.Seconds(), "sec")
			time.Sleep(c.poll)
		}
	}()

	go func() {
		for {
			log.Println("[REPORT]Wait for", c.report.Seconds(), "sec")
			time.Sleep(c.report)
			log.Println("[REPORT]Call the handler")
			c.rh(c.getData())
		}
	}()

}

type Metric struct {
	Name, Value string
	Type        server.MetricType
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

func getMemMetrics() []Metric {
	ret := make([]Metric, 0, len(memMetrics))
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
			ret = append(ret, Metric{Name: name, Type: server.Gauge, Value: statValue})
		}
	}
	return ret
}

func getExtraMetrics() []Metric {
	ret := make([]Metric, 0, 2)

	ret = append(
		ret, Metric{
			Name:  "RandomValue",
			Type:  server.Gauge,
			Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
		},
	)

	ret = append(
		ret,
		Metric{
			Name:  "PollCount",
			Type:  server.Counter,
			Value: strconv.Itoa(count()),
		},
	)

	return ret
}
