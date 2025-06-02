package provider

import (
	"time"

	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

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

type StatProvider struct {
	poll      time.Duration
	providers []di.MetricsProvider
}

func New(poll time.Duration) *StatProvider {
	p := &StatProvider{poll: poll}
	p.providers = append(
		p.providers,
		newManualStat(poll), newPSUtilStat(poll), newRuntimeStat(poll),
	)
	return p
}

func (p *StatProvider) Run() {
	for _, v := range p.providers {
		go v.Run()
	}
}

func (p *StatProvider) Stop() {
	for _, v := range p.providers {
		v.Stop()
	}
}

func (p *StatProvider) GetMetrics() metrics.MetricsList {
	var m metrics.MetricsList
	for _, p := range p.providers {
		m = append(m, p.GetMetrics()...)
	}
	return m
}
