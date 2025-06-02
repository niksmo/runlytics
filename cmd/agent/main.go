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
}
