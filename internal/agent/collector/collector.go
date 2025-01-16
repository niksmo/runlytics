package collector

import (
	"log"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/niksmo/runlytics/pkg/counter"
)

type metricsData struct {
	counter map[string]int64
	gauge   map[string]float64
}

type collector struct {
	poll    time.Duration
	data    metricsData
	counter *counter.Counter
	mu      sync.Mutex
}

func New(interval time.Duration) *collector {
	c := &collector{
		poll:    interval,
		data:    metricsData{make(map[string]int64), make(map[string]float64)},
		counter: counter.New(),
	}

	return c
}

func (c *collector) Run() {
	log.Printf(
		"Run metrics collector with poll interval = %vs\n",
		c.poll.Seconds(),
	)

	for {
		c.collectMetrics()
		log.Println("[POLL]Wait for", c.poll.Seconds(), "sec")
		time.Sleep(c.poll)
	}

}

func (c *collector) GetGaugeMetrics() map[string]float64 {
	ret := make(map[string]float64, len(c.data.gauge))
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.data.gauge {
		ret[k] = v
	}

	return ret
}

func (c *collector) GetCounterMetrics() map[string]int64 {
	ret := make(map[string]int64, len(c.data.gauge))
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.data.counter {
		ret[k] = v
	}

	return ret
}

func (c *collector) collectMetrics() {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)

	c.mu.Lock()
	defer c.mu.Unlock()

	// memory
	c.data.gauge["Alloc"] = float64(memStat.Alloc)
	c.data.gauge["BuckHashSys"] = float64(memStat.BuckHashSys)
	c.data.gauge["Frees"] = float64(memStat.Frees)
	c.data.gauge["GCCPUFraction"] = memStat.GCCPUFraction
	c.data.gauge["GCSys"] = float64(memStat.GCSys)
	c.data.gauge["HeapAlloc"] = float64(memStat.HeapAlloc)
	c.data.gauge["HeapIdle"] = float64(memStat.HeapIdle)
	c.data.gauge["HeapInuse"] = float64(memStat.HeapInuse)
	c.data.gauge["HeapObjects"] = float64(memStat.HeapObjects)
	c.data.gauge["HeapReleased"] = float64(memStat.HeapReleased)
	c.data.gauge["HeapSys"] = float64(memStat.HeapSys)
	c.data.gauge["LastGC"] = float64(memStat.LastGC)
	c.data.gauge["Lookups"] = float64(memStat.Lookups)
	c.data.gauge["MCacheInuse"] = float64(memStat.MCacheInuse)
	c.data.gauge["MCacheSys"] = float64(memStat.MCacheSys)
	c.data.gauge["MSpanInuse"] = float64(memStat.MSpanInuse)
	c.data.gauge["MSpanSys"] = float64(memStat.MSpanSys)
	c.data.gauge["Mallocs"] = float64(memStat.Mallocs)
	c.data.gauge["NextGC"] = float64(memStat.NextGC)
	c.data.gauge["NumForcedGC"] = float64(memStat.NumForcedGC)
	c.data.gauge["NumGC"] = float64(memStat.NumGC)
	c.data.gauge["OtherSys"] = float64(memStat.OtherSys)
	c.data.gauge["PauseTotalNs"] = float64(memStat.PauseTotalNs)
	c.data.gauge["StackInuse"] = float64(memStat.StackInuse)
	c.data.gauge["StackSys"] = float64(memStat.StackSys)
	c.data.gauge["Sys"] = float64(memStat.Sys)
	c.data.gauge["TotalAlloc"] = float64(memStat.TotalAlloc)

	// extra
	c.data.gauge["RandomValue"] = rand.Float64()
	c.data.counter["PollCount"] = c.counter.Next()
}
