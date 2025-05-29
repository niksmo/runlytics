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
	config := config.Load()
	logger.Init(config.Log.Level)
	config.PrintConfig(logger.Log)

	application := app.New(config)
	go application.HTTPServer.MustRun()
	go application.GRPCServer.MustRun()
	go application.Storage.MustRun()

	stopCtx, stopFn := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stopFn()

	<-stopCtx.Done()
	application.HTTPServer.Stop()
	application.GRPCServer.Stop()
	application.Storage.Stop()
}
