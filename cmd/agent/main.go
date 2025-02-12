package main

import (
	"net/http"

	"github.com/niksmo/runlytics/internal/agent/collector"
	"github.com/niksmo/runlytics/internal/agent/emitter"
	"github.com/niksmo/runlytics/internal/logger"
)

func main() {
	parseFlags()

	logger.Init(flagLog)

	logger.Log.Debug("Start agent")

	collector := collector.New(flagPoll)

	HTTPEmitter := emitter.New(
		flagReport,
		collector,
		http.DefaultClient,
		flagAddr,
	)

	go collector.Run()
	HTTPEmitter.Run()
}
