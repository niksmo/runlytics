package reportgen

import (
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
	"go.uber.org/zap"
)

type ReportGen struct {
	provider di.MetricsProvider
	report   time.Duration
	c        chan metrics.MetricsList
	C        <-chan metrics.MetricsList
	ticker   *time.Ticker
}

func New(p di.MetricsProvider, report time.Duration) *ReportGen {
	c := make(chan metrics.MetricsList, 1)
	return &ReportGen{
		provider: p,
		report:   report,
		c:        c,
		C:        c,
		ticker:   time.NewTicker(report),
	}
}

func (g *ReportGen) Run() {
	const op = "reportgen.Run"
	log := logger.Log.With(
		zap.String("op", op), zap.Duration("reportInt", g.report),
	)
	log.Info("running")
	for range g.ticker.C {
		t := time.Now()
		g.c <- g.provider.GetMetrics()
		log.Debug("send metrics to channel", zap.Duration("waitConsumers", time.Since(t)))
	}
}

func (g *ReportGen) Stop() {
	const op = "reportgen.Stop"
	g.ticker.Stop()
	close(g.c)
	logger.Log.Info("stopped", zap.String("op", op))
}
