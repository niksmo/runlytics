package generator

import (
	"context"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/counter"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type Job struct {
	payload []metrics.Metrics
	id      int64
}

func (job *Job) ID() int64 {
	return job.id
}

func (job *Job) Payload() []metrics.Metrics {
	return job.payload
}

type JobGenerator struct {
	counter *counter.Counter
	pollCounter
	interval time.Duration
}

type pollCounter struct {
	rollbackValue func()
	jobID         int64
	prevValue     int64
}

func New(interval time.Duration) *JobGenerator {
	return &JobGenerator{
		interval: interval,
		counter:  counter.New(),
	}
}

func (g *JobGenerator) Run(
	stopCtx context.Context,
	jobCh chan<- di.IJob,
	errCh <-chan di.IJobErr,
	collectors []di.IMetricsCollector,
) {
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()
	defer close(jobCh)
	for {
		logger.Log.Debug("JobGenerator wait", zap.Float64("seconds", g.interval.Seconds()))
		select {
		case <-stopCtx.Done():
			logger.Log.Debug("JobGenerator close jobStream")
			return
		case <-ticker.C:
			for len(errCh) != 0 {
				err := <-errCh
				if err.ID() == g.pollCounter.jobID {
					g.pollCounter.rollbackValue()
				}
			}
			for _, collector := range collectors {
				jobCh <- g.makeJob(g.getJobID(), collector)
			}
		}

	}
}

func (g *JobGenerator) getJobID() int64 {
	return g.counter.Next()
}

func (g *JobGenerator) makeJob(
	id int64, collector di.IMetricsCollector,
) *Job {
	var payload []metrics.Metrics
	for name, value := range collector.GetGaugeMetrics() {
		payload = append(
			payload,
			metrics.Metrics{
				ID: name, MType: metrics.MTypeGauge, Value: &value,
			},
		)
	}

	for name, value := range collector.GetCounterMetrics() {
		if name == "PollCount" {
			prev := g.pollCounter.prevValue
			g.pollCounter.prevValue = value
			value = value - prev
			g.pollCounter.jobID = id
			g.pollCounter.rollbackValue = func() {
				g.pollCounter.prevValue = prev
				logger.Log.Debug("Rollback pollCount prevValue", zap.Int64("actual", prev))
			}
		}
		payload = append(
			payload,
			metrics.Metrics{
				ID: name, MType: metrics.MTypeCounter, Delta: &value,
			},
		)
	}
	return &Job{id: id, payload: payload}
}
