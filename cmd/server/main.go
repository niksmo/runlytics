package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/niksmo/runlytics/internal/buildinfo"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/app"
	"github.com/niksmo/runlytics/internal/server/config"
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
	go application.HTTPServer.MustRun()
	go application.GRPCServer.MustRun()
	go application.Storage.MustRun()

	<-stopCtx.Done()
	application.HTTPServer.Stop()
	application.GRPCServer.Stop()
	application.Storage.Stop()
}
