package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/niksmo/runlytics/internal/agent/app"
	"github.com/niksmo/runlytics/internal/agent/config"
	"github.com/niksmo/runlytics/internal/buildinfo"
	"github.com/niksmo/runlytics/internal/logger"
)

func main() {
	buildinfo.Print()
	cfg := config.Load()
	logger.Init(cfg.Log.Level)
	cfg.PrintConfig(logger.Log)

	stopCtx, stopFn := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stopFn()

	application := app.New(cfg)
	go application.Provider.Run()
	go application.ReportGen.Run()
	go application.WorkerPool.Run()

	<-stopCtx.Done()
	application.Provider.Stop()
	application.ReportGen.Stop()
	application.WorkerPool.Stop()

	// collectors := runCollectors(cfg)
	// jobCh, errCh := runJobGenerator(stopCtx, collectors, cfg)

	// encrypter, err := cipher.NewEncrypterX509(cfg.Crypto.Data)
	// if err != nil {
	// 	logger.Log.Fatal("failed to init encrypter", zap.Error(err))
	// }
	// var wg sync.WaitGroup
	// workerParams := worker.WorkerParams{
	// 	Wg:         &wg,
	// 	JobCh:      jobCh,
	// 	ErrCh:      errCh,
	// 	URL:        cfg.Server.URL(),
	// 	Key:        cfg.HashKey.Key,
	// 	Encrypter:  encrypter,
	// 	HTTPClient: &http.Client{Timeout: cfg.HTTPClientTimeout()},
	// 	OutboundIP: cfg.GetOutboundIP(),
	// }

	// runWorkers(
	// 	workerParams,
	// 	cfg,
	// 	logger.Log,
	// )

	// <-stopCtx.Done()
	// logger.Log.Info("shutdown garecefully")
	// wg.Wait()
	// logger.Log.Info("workers stopped")
}

// func runCollectors(config *config.AgentConfig) []di.IMetricsCollector {
// 	collectors := []di.IMetricsCollector{
// 		collector.NewRuntimeMemStat(config.Metrics.Poll),
// 		collector.NewManualStat(config.Metrics.Poll),
// 		collector.NewPsUtilStat(config.Metrics.Poll),
// 	}

// 	for _, collector := range collectors {
// 		go collector.Run()
// 	}
// 	return collectors
// }

// func runJobGenerator(
// 	ctx context.Context,
// 	collectors []di.IMetricsCollector,
// 	config *config.AgentConfig,
// ) (jobCh chan di.IJob, errCh chan di.IJobErr) {
// 	jobCh = make(chan di.IJob, config.Metrics.JobsBuf)
// 	errCh = make(chan di.IJobErr, config.Metrics.JobsErrBuf)
// 	jobGenerator := generator.New(config.Metrics.Report)
// 	go jobGenerator.Run(ctx, jobCh, errCh, collectors)
// 	return jobCh, errCh
// }

// func runWorkers(
// 	params worker.WorkerParams, config *config.AgentConfig, logger *zap.Logger,
// ) {
// 	for idx := range config.Metrics.RateLimit {
// 		go worker.Run(params)
// 		logger.Info("Worker is running", zap.Int("workerIdx", idx))
// 	}
// }
