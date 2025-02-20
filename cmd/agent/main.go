package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/niksmo/runlytics/internal/agent/collector"
	"github.com/niksmo/runlytics/internal/agent/config"
	"github.com/niksmo/runlytics/internal/agent/generator"
	"github.com/niksmo/runlytics/internal/agent/worker"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

func main() {
	config := config.Load()

	logger.Init(config.LogLvl())
	logger.Log.Info(
		"Start agent with flags",
		zap.String("ADDRESS", config.Addr().String()),
		zap.String("LOG_LVL", config.LogLvl()),
		zap.String("POLL_INTERVAL", config.Poll().String()),
		zap.String("REPORT_INTERVAL", config.Report().String()),
		zap.String("KEY", config.Key()),
		zap.Int("RATE_LIMIT", config.RateLimit()),
	)

	stopCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	HTTPClient := &http.Client{Timeout: config.Report() - 100*time.Millisecond}
	URL := config.Addr().JoinPath("updates").String()
	jobCh := make(chan di.Job, 1024)
	errCh := make(chan di.JobErr, 128)
	jobGenerator := generator.New(config.Report())

	collectors := []di.MetricsCollector{
		collector.NewRuntimeMemStat(config.Poll()),
		collector.NewManualStat(config.Poll()),
	}
	for _, collector := range collectors {
		go collector.Run()
	}

	go jobGenerator.Run(jobCh, errCh, collectors)

	for idx := range config.RateLimit() {
		go worker.Run(jobCh, errCh, URL, config.Key(), HTTPClient)
		logger.Log.Info("Worker is running", zap.Int("workerIdx", idx))
	}
	<-stopCtx.Done()
}
