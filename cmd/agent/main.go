package main

import (
	"fmt"
	"log"
	"time"

	"github.com/niksmo/runlytics/internal/agent"
)

const pollInterval = time.Duration(2 * time.Second)
const reportInterval = time.Duration(10 * time.Second)

func main() {
	log.Println("Start agent")
	collector := agent.NewCollector(
		pollInterval,
		reportInterval,
		func(data map[string]agent.Metric) {
			log.Println("[HANDLER]: Reporting", time.Now())
			for name, metric := range data {
				fmt.Println(name, metric.T, metric.V)
			}
		},
	)

	collector.Run()
}
