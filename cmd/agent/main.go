package main

import (
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
		func() { log.Println("Report", time.Now()) },
	)

	collector.Run()
}
