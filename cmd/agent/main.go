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
	logger.Init(config.LogLvl())

	printAgentConfig(config, logger.Log)

	encrypter, err := cipher.NewEncrypter(config.CryptoKeyData())
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
		URL:        config.Addr().JoinPath("updates").String(),
		Key:        config.Key(),
		Encrypter:  encrypter,
		HTTPClient: &http.Client{Timeout: config.HTTPClientTimeout()},
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

func printAgentConfig(config *config.Config, logger *zap.Logger) {
	logger.Info(
		"Start agent with flags",
		zap.String("ADDRESS", config.Addr().String()),
		zap.String("LOG_LVL", config.LogLvl()),
		zap.String("POLL_INTERVAL", config.Poll().String()),
		zap.String("REPORT_INTERVAL", config.Report().String()),
		zap.String("KEY", config.Key()),
		zap.Int("RATE_LIMIT", config.RateLimit()),
		zap.String("CRYPTO_KEY", config.CryptoKeyPath()),
	)
}

func runCollectors(config *config.Config) []di.MetricsCollector {
	collectors := []di.MetricsCollector{
		collector.NewRuntimeMemStat(config.Poll()),
		collector.NewManualStat(config.Poll()),
		collector.NewPsUtilStat(config.Poll()),
	}

	for _, collector := range collectors {
		go collector.Run()
	}
	return collectors
}

func runJobGenerator(
	ctx context.Context,
	collectors []di.MetricsCollector,
	config *config.Config,
) (jobCh chan di.Job, errCh chan di.JobErr) {
	jobCh = make(chan di.Job, config.JobsBuf())
	errCh = make(chan di.JobErr, config.JobsErrBuf())
	jobGenerator := generator.New(config.Report())
	go jobGenerator.Run(ctx, jobCh, errCh, collectors)
	return jobCh, errCh
}

func runWorkers(
	params worker.WorkerParams, config *config.Config, logger *zap.Logger,
) {
	for idx := range config.RateLimit() {
		go worker.Run(params)
		logger.Info("Worker is running", zap.Int("workerIdx", idx))
	}
}
