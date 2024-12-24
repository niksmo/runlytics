package agent

import (
	"log"
	"sync"
	"time"
)

type Collector struct {
	pollInt, reportInt time.Duration
	handler            func(data map[string]Metric)
}

func NewCollector(pollInt, reropInt time.Duration, handler func(data map[string]Metric)) *Collector {
	return &Collector{pollInt, reropInt, handler}
}

func (c *Collector) Run() {
	log.Printf(
		"Run collector with intervals: poll = %vs, report = %vs\n",
		c.pollInt.Seconds(), c.reportInt.Seconds(),
	)

	var wg sync.WaitGroup

	m := &Metrics{data: make(map[string]Metric)}

	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			log.Println("[POLL]Start collect")
			m.mu.Lock()
			getMemMetrics(m.data)
			getExtraMetrics(m.data)
			m.mu.Unlock()
			log.Println("[POLL]Stop collect")
			log.Println("[POLL]Wait for", c.pollInt.Seconds(), "sec")
			time.Sleep(c.pollInt)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			log.Println("[REPORT]Wait for", c.reportInt.Seconds(), "sec")
			time.Sleep(c.reportInt)
			log.Println("[REPORT]Call the handler")
			m.mu.Lock()
			c.handler(m.data)
			m.mu.Unlock()
		}
	}()
	wg.Wait()
}
