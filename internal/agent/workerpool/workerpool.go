package workerpool

import (
	"context"
	"fmt"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type WorkerOpts struct {
	URL        string
	HashKey    string
	Encrypter  di.Encrypter
	OutboundIP string
}

type WorkerPool struct {
	n         int
	in        <-chan metrics.MetricsList
	pollCount [2]int64
	wf        di.SendMetricsFunc
	wo        WorkerOpts
	grp       *errgroup.Group
}

func New(
	n int, in <-chan metrics.MetricsList, wf di.SendMetricsFunc, wo WorkerOpts,
) *WorkerPool {
	return &WorkerPool{
		n:  n,
		in: in,
		wf: wf,
		wo: wo,
	}
}

func (p *WorkerPool) Run() {
	const op = "workerpool.Run"
	for m := range p.in {
		pci, ok := p.findPollCountIdx(m)

		if ok {
			p.handlePollCount(pci, m)
		}

		output := p.divideInput(m)

		if err := p.doWork(output); err != nil {
			logger.Log.Warn(
				"failed to do work", zap.String("op", op), zap.Error(err),
			)
			p.rollbackPollCount()
		}
	}
}

func (p *WorkerPool) Stop() {
	p.grp.Wait()
}

func (p *WorkerPool) findPollCountIdx(m metrics.MetricsList) (int, bool) {
	for i, v := range m {
		if v.ID == "PollCount" {
			return i, true
		}
	}
	return -1, false
}

func (p *WorkerPool) handlePollCount(idx int, m metrics.MetricsList) {
	p.pollCount[0] = p.pollCount[1]
	d := m[idx].Delta - p.pollCount[1]
	p.pollCount[1] = m[idx].Delta
	m[idx].Delta = d
}

func (p *WorkerPool) rollbackPollCount() {
	const op = "workerpool.rollbackPollCount"
	p.pollCount[1] = p.pollCount[0]
	logger.Log.Debug(
		"rollback",
		zap.String("op", op), zap.Int64s("pollCount", p.pollCount[:]),
	)
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
	const op = "workerpool.doWork"
	grp, ctx := errgroup.WithContext(context.Background())
	p.grp = grp

	for _, m := range output {
		ml := m
		grp.Go(func() error {
			return p.wf(
				ctx,
				ml,
				p.wo.Encrypter,
				p.wo.URL,
				p.wo.HashKey,
				p.wo.OutboundIP,
			)
		})
	}

	if err := grp.Wait(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
