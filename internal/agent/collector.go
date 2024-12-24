package agent

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrReportIntLessPollInt = errors.New("report interval should be more or equal to poll interval")
	ErrMinIntervalValue     = errors.New("both intervals should be more or equal 1s")
)

type Collector struct {
	poll, report time.Duration
	rh           ReportHandler
}

type ReportHandler func(data map[string]Metric)

func NewCollector(poll, report time.Duration, rh ReportHandler) (*Collector, error) {
	s := time.Duration(1 * time.Second)
	if poll < s || report < s {
		return nil, ErrMinIntervalValue
	}

	if report < poll {
		return nil, ErrReportIntLessPollInt
	}

	return &Collector{poll, report, rh}, nil
}

func (c *Collector) Run() {
	log.Printf(
		"Run collector with intervals: poll = %vs, report = %vs\n",
		c.poll.Seconds(), c.report.Seconds(),
	)

	var wg sync.WaitGroup

	m := &Metrics{data: make(map[string]Metric)}

	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			m.mu.Lock()
			getMemMetrics(m.data)
			getExtraMetrics(m.data)
			m.mu.Unlock()
			log.Println("[POLL]Wait for", c.poll.Seconds(), "sec")
			time.Sleep(c.poll)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			log.Println("[REPORT]Wait for", c.report.Seconds(), "sec")
			time.Sleep(c.report)
			log.Println("[REPORT]Call the handler")
			m.mu.Lock()
			c.rh(m.data)
			m.mu.Unlock()
		}
	}()
	wg.Wait()
}
