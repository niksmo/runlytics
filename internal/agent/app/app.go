package app

import (
	"github.com/niksmo/runlytics/internal/agent/config"
	"github.com/niksmo/runlytics/internal/agent/generator"
	"github.com/niksmo/runlytics/internal/agent/provider"
	"github.com/niksmo/runlytics/internal/agent/workerpool"
	grpcworker "github.com/niksmo/runlytics/internal/agent/workerpool/grpc"
	"github.com/niksmo/runlytics/pkg/di"
)

type App struct {
	Provider   di.RunStopper
	ReportGen  di.RunStopper
	WorkerPool di.RunStopper
}

func New(cfg *config.AgentConfig) *App {
	provider := provider.New(cfg.Metrics.Poll)
	reportGen := generator.NewReportGen(provider, cfg.Metrics.Report)
	wPool := workerpool.New(
		cfg.Metrics.RateLimit, grpcworker.SendMetrics, reportGen.C,
	)

	return &App{
		Provider:   provider,
		ReportGen:  reportGen,
		WorkerPool: wPool,
	}
}
