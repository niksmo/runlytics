package main

import (
	"log"
	"strconv"
	"time"

	"github.com/niksmo/runlytics/internal/agent"
)

const (
	poll        = time.Duration(2 * time.Second)
	report      = time.Duration(10 * time.Second)
	defaultHost = "http://127.0.0.1"
	defaultPort = 8080
)

func main() {
	log.Println("Start agent")
	collector, err := agent.NewCollector(
		poll,
		report,
		handler(),
	)

	if err != nil {
		log.Fatal(err)
	}

	collector.Run()
}

func handler() agent.ReportHandler {
	addr := defaultHost + ":" + strconv.Itoa(defaultPort)
	httpEmitter, err := agent.NewHttpEmitter(addr)
	if err != nil {
		log.Fatal(err)
	}

	return func(data map[string]agent.Metric) {
		log.Println("[HANDLER]: Reporting")
		for name, metric := range data {
			httpEmitter(metric.T, name, metric.V)
		}
	}
}
