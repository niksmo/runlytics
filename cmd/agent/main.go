package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	go collector.Run()
	wg.Wait()
}

func handler() agent.ReportHandler {
	addr := defaultHost + ":" + strconv.Itoa(defaultPort)
	httpEmittingFunc, err := agent.HTTPEmittingFunc(addr, http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	return func(data []agent.Metric) {
		log.Println("[HANDLER]: Reporting")
		for _, metric := range data {
			httpEmittingFunc(string(metric.Type), metric.Name, metric.Value)
		}
	}
}
