package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/niksmo/runlytics/internal/agent"
)

func main() {
	parseFlags()
	log.Println("Start agent")
	collector, err := agent.NewCollector(
		time.Duration(flagPoll)*time.Second,
		time.Duration(flagReport)*time.Second,
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
	httpEmittingFunc, err := agent.HTTPEmittingFunc(fmt.Sprintf("http://%s", flagAddr), http.DefaultClient)
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
