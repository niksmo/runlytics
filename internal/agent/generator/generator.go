package generator

import (
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/counter"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type Job struct {
	id      int64
	payload []metrics.MetricsUpdate
}

func (job *Job) Payload() []metrics.MetricsUpdate {
	return job.payload
}
func (job *Job) ID() int64 {
	return job.id
}

type JobGenerator struct {
	interval time.Duration
	counter  *counter.Counter
	pollCounter
}

type pollCounter struct {
	jobID         int64
	prevValue     int64
	rollbackValue func()
}

func New(interval time.Duration) *JobGenerator {
	return &JobGenerator{
		interval: interval,
		counter:  counter.New(),
	}
}

func (g *JobGenerator) Run(
	jobCh chan<- di.Job,
	errCh <-chan di.JobErr,
	collectors []di.MetricsCollector,
) {
	defer close(jobCh)
	for {
		logger.Log.Debug("Generator wait", zap.Float64("seconds", g.interval.Seconds()))
		time.Sleep(g.interval)
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

func (g *JobGenerator) getJobID() int64 {
	return g.counter.Next()
}

func (g *JobGenerator) makeJob(
	id int64, collector di.MetricsCollector,
) di.Job {
	var payload []metrics.MetricsUpdate
	for name, value := range collector.GetGaugeMetrics() {
		payload = append(
			payload,
			metrics.MetricsUpdate{
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
			metrics.MetricsUpdate{
				ID: name, MType: metrics.MTypeCounter, Delta: &value,
			},
		)
	}
	return &Job{id: id, payload: payload}
}
