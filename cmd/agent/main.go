package main

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/niksmo/runlytics/internal/agent/collector"
	"github.com/niksmo/runlytics/internal/agent/config"
	"github.com/niksmo/runlytics/internal/agent/generator"
	"github.com/niksmo/runlytics/internal/agent/worker"
	"github.com/niksmo/runlytics/internal/buildinfo"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/cipher"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

func main() {
	stopCtx, stopFn := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stopFn()

	buildinfo.Print()

	config := config.Load()
	logger.Init(config.Log.Level)

	printAgentConfig(config, logger.Log)

	encrypter, err := cipher.NewEncrypterX509(config.Crypto.Data)
	if err != nil {
		logger.Log.Fatal("failed to init encrypter", zap.Error(err))
	}

	collectors := runCollectors(config)
	jobCh, errCh := runJobGenerator(stopCtx, collectors, config)

	var wg sync.WaitGroup
	workerParams := worker.WorkerParams{
		Wg:         &wg,
		JobCh:      jobCh,
		ErrCh:      errCh,
		URL:        config.Server.URL(),
		Key:        config.HashKey.Key,
		Encrypter:  encrypter,
		HTTPClient: &http.Client{Timeout: config.HTTPClientTimeout()},
		OutboundIP: config.GetOutboundIP(),
	}

	runWorkers(
		workerParams,
		config,
		logger.Log,
	)

	<-stopCtx.Done()
	logger.Log.Info("shutdown garecefully")
	wg.Wait()
	logger.Log.Info("workers stopped")
}

func printAgentConfig(config *config.AgentConfig, logger *zap.Logger) {
	logger.Info(
		"Start agent with flags",
		zap.String("ADDRESS", config.Server.URL()),
		zap.String("LOG_LVL", config.Log.Level),
		zap.String("POLL_INTERVAL", config.Metrics.Poll.String()),
		zap.String("REPORT_INTERVAL", config.Metrics.Report.String()),
		zap.String("KEY", config.HashKey.Key),
		zap.Int("RATE_LIMIT", config.Metrics.RateLimit),
		zap.String("CRYPTO_KEY", config.Crypto.Path),
		zap.String("OUTBOUND_IP", config.GetOutboundIP()),
	)
}

func runCollectors(config *config.AgentConfig) []di.MetricsCollector {
	collectors := []di.MetricsCollector{
		collector.NewRuntimeMemStat(config.Metrics.Poll),
		collector.NewManualStat(config.Metrics.Poll),
		collector.NewPsUtilStat(config.Metrics.Poll),
	}

	for _, collector := range collectors {
		go collector.Run()
	}
	return collectors
}

func runJobGenerator(
	ctx context.Context,
	collectors []di.MetricsCollector,
	config *config.AgentConfig,
) (jobCh chan di.Job, errCh chan di.JobErr) {
	jobCh = make(chan di.Job, config.Metrics.JobsBuf)
	errCh = make(chan di.JobErr, config.Metrics.JobsErrBuf)
	jobGenerator := generator.New(config.Metrics.Report)
	go jobGenerator.Run(ctx, jobCh, errCh, collectors)
	return jobCh, errCh
}

func runWorkers(
	params worker.WorkerParams, config *config.AgentConfig, logger *zap.Logger,
) {
	for idx := range config.Metrics.RateLimit {
		go worker.Run(params)
		logger.Info("Worker is running", zap.Int("workerIdx", idx))
	}
}
