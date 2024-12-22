package main

import (
	"fmt"
	"runtime"

	"github.com/niksmo/runlytics/pkg/field"
)

var memMetrics = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

func main() {
	memStat := new(runtime.MemStats)

	runtime.ReadMemStats(memStat)

	for _, metric := range memMetrics {
		alloc, err := field.Value(memStat, metric)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%v: %v\n", metric, alloc)
	}
}
