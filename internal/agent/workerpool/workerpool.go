package workerpool

import (
	"context"

	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"golang.org/x/sync/errgroup"
)

type WorkerPool struct {
	n         int
	in        <-chan metrics.MetricsList
	w         di.SendMetricsFunc
	pollCount [2]int64
}

func New(n int, w di.SendMetricsFunc, in <-chan metrics.MetricsList) *WorkerPool {
	return &WorkerPool{
		n:  n,
		in: in,
		w:  w,
	}
}

func (p *WorkerPool) Run() {
	for m := range p.in {
		pcm, ok := p.findPollCount(m)

		if ok {
			p.handlePollCount(pcm)
		}

		output := p.divideInput(m)

		if err := p.doWork(output); err != nil {
			p.rollbackPollCount()
		}
	}
}

func (p *WorkerPool) Stop() {}

func (p *WorkerPool) findPollCount(m metrics.MetricsList) (metrics.Metrics, bool) {
	for _, v := range m {
		if v.ID == "PollCount" {
			return v, true
		}
	}
	return metrics.Metrics{}, false
}

func (p *WorkerPool) handlePollCount(pcm metrics.Metrics) {
	p.pollCount[0] = p.pollCount[1]
	d := *pcm.Delta - p.pollCount[1]
	p.pollCount[1] = *pcm.Delta
	pcm.Delta = &d
}

func (p *WorkerPool) rollbackPollCount() {
	p.pollCount[1] = p.pollCount[0]
}

func (p *WorkerPool) divideInput(m metrics.MetricsList) []metrics.MetricsList {
	var output []metrics.MetricsList
	if len(m) > p.n {
		chunkSize := len(m) / p.n
		lastIdx := p.n - 1
		for idx := range p.n {
			if idx == lastIdx {
				output = append(output, m)
			} else {
				output = append(output, m[:chunkSize])
				m = m[chunkSize:]
			}
		}
	} else {
		for idx := range m {
			output = append(output, m[idx:idx+1])
		}
	}
	return output
}

func (p *WorkerPool) doWork(output []metrics.MetricsList) error {
	grp, ctx := errgroup.WithContext(context.Background())

	for _, m := range output {
		ml := m
		grp.Go(func() error {
			p.w(ctx, ml)
			return nil
		})
	}

	return grp.Wait()
}
