package app

import (
	"github.com/niksmo/runlytics/internal/agent/config"
	"github.com/niksmo/runlytics/internal/agent/provider"
	"github.com/niksmo/runlytics/internal/agent/reportgen"
	"github.com/niksmo/runlytics/internal/agent/workerpool"
	grpcworker "github.com/niksmo/runlytics/internal/agent/workerpool/grpc"
	httpworker "github.com/niksmo/runlytics/internal/agent/workerpool/http"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/cipher"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

type App struct {
	Provider   di.RunStopper
	ReportGen  di.RunStopper
	WorkerPool di.RunStopper
}

func New(cfg *config.AgentConfig) *App {
	provider := provider.New(cfg.Metrics.Poll)
	reportGen := reportgen.New(provider, cfg.Metrics.Report)
	encrypter, err := cipher.NewEncrypterX509(cfg.Crypto.Data)
	if err != nil {
		logger.Log.Fatal("failed to init encrypter", zap.Error(err))
	}
	wo := workerpool.WorkerOpts{
		Encrypter:  encrypter,
		URL:        cfg.Server.URL(),
		HashKey:    cfg.HashKey.Key,
		OutboundIP: cfg.GetOutboundIP(),
	}

	var wf di.SendMetricsFunc
	if cfg.GRPC.IsSet {
		wf = grpcworker.SendMetrics
	} else {
		wf = httpworker.SendMetrics
	}
	wPool := workerpool.New(
		cfg.Metrics.RateLimit, reportGen.C, wf, wo,
	)

	return &App{
		Provider:   provider,
		ReportGen:  reportGen,
		WorkerPool: wPool,
	}
}
