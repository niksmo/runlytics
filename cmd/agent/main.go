package main

import (
	"log"
	"net/http"

	"github.com/niksmo/runlytics/internal/agent/collector"
	"github.com/niksmo/runlytics/internal/agent/emitter"
)

func main() {
	parseFlags()

	log.Println("Start agent")

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
