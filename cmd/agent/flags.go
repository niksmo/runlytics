package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	flagAddr   *addr = &addr{host: "localhost", port: 8080}
	flagPoll   int
	flagReport int
)

type addr struct {
	host string
	port int
}

func (a *addr) String() string {
	return a.host + ":" + strconv.Itoa(a.port)
}

func (a *addr) Set(v string) error {
	var errAddr = errors.New("incorrect address format, usage: example.com:8080")
	slice := strings.SplitN(v, ":", 2)
	if len(slice) != 2 {
		return errAddr
	}

	a.host = slice[0]

	port, err := strconv.Atoi(slice[1])

	if err != nil {
		return errAddr
	}

	a.port = port
	return nil
}

func parseFlags() {
	flag.Var(
		flagAddr,
		"a",
		"Host address for metrics emitting, e.g. example.com:8080",
	)

	flag.IntVar(&flagPoll,
		"p",
		2,
		"Polling collecting metrics interval in sec, e.g. 5",
	)

	flag.IntVar(&flagReport,
		"r",
		10,
		"Emitting metrics interval in sec, e.g. 10",
	)

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		if err := flagAddr.Set(envAddr); err != nil {
			log.Print(fmt.Errorf("parse env ADDRESS error: %w", err))
		}
	}

	getIntervalEnv(&flagPoll, "POLL_INTERVAL")
	getIntervalEnv(&flagReport, "REPORT_INTERVAL")
}

func getIntervalEnv(interval *int, param string) {
	if envValue := os.Getenv(param); envValue != "" {
		reportInt, err := strconv.Atoi(envValue)
		if err != nil {
			log.Print(fmt.Errorf("parse env %s error: %w", param, err))
		} else {
			*interval = reportInt
		}
	}
}
