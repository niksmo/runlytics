package main

import (
	"net/http"

	"github.com/niksmo/runlytics/internal/agent/collector"
	"github.com/niksmo/runlytics/internal/agent/config"
	"github.com/niksmo/runlytics/internal/agent/emitter"
	"github.com/niksmo/runlytics/internal/logger"
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
	)

	collector := collector.New(config.Poll())

	HTTPEmitter := emitter.New(
		config.Report(),
		collector,
		http.DefaultClient,
		config.Addr(),
	)

	go collector.Run()
	HTTPEmitter.Run()
}
