package main

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	minPollInterval   = time.Second
	minReportInterval = time.Second
)

var (
	ErrReportLessPoll    = errors.New("report interval should be more or equal to poll interval")
	ErrMinPollInterval   = errors.New("polling interval should be more or equal 1s")
	ErrMinReportInterval = errors.New("report interval should be more or equal 1s")
	ErrAddrParse         = errors.New("used unnecessary url symbols in address")
)

type flagsErrors []error

func (fe flagsErrors) Error() string {
	s := make([]string, 0, len(fe))
	for _, e := range fe {
		s = append(s, e.Error())
	}
	return strings.Join(s, ", ")
}

var (
	flagAddr   *url.URL
	flagPoll   time.Duration
	flagReport time.Duration
	flagLog    string
)

func parseFlags() {
	addr := flag.String(
		"a",
		"localhost:8080",
		"Host address for metrics emitting, e.g. example.com:8080")

	poll := flag.Int(
		"p",
		2,
		"Polling collecting metrics interval in sec, e.g. 5",
	)

	report := flag.Int(
		"r",
		10,
		"Emitting metrics interval in sec, e.g. 10",
	)

	flag.StringVar(&flagLog, "l", "info", "Set level, e.g. debug")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = &envAddr
	}

	err := getIntervalEnv(poll, "POLL_INTERVAL")
	if err != nil {
		log.Printf("Error on parse POLL_INTERVAL env: %v", err)
		log.Printf("Current poll flag value is: %v sec\n", flagPoll.Seconds())
	}

	err = getIntervalEnv(report, "REPORT_INTERVAL")
	if err != nil {
		log.Printf("Error on parse REPORT_INTERVAL env: %v", err)
		log.Printf("Current report flag value is: %v sec\n", flagPoll.Seconds())
	}

	if envLog := os.Getenv("LOG_LVL"); envLog != "" {
		flagLog = envLog
	}

	flagPoll = time.Duration(*poll) * time.Second
	flagReport = time.Duration(*report) * time.Second

	var fatal bool

	err = checkIntervals()
	if err != nil {
		fatal = true
		log.Println(err)
	}

	err = setFlagAddr(*addr)
	if err != nil {
		fatal = true
		log.Println(err)
	}

	if fatal {
		os.Exit(1)
	}
}

func getIntervalEnv(interval *int, param string) error {
	if envValue := os.Getenv(param); envValue != "" {
		reportInt, err := strconv.Atoi(envValue)
		if err != nil {
			return err
		}

		*interval = reportInt
	}
	return nil
}

func checkIntervals() error {
	var errs flagsErrors

	if flagPoll < minPollInterval {
		errs = append(errs, ErrMinPollInterval)
	}

	if flagReport < minReportInterval {
		errs = append(errs, ErrMinReportInterval)
	}

	if flagReport < flagPoll {
		errs = append(errs, ErrReportLessPoll)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func setFlagAddr(addr string) error {
	if !(strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")) {
		addr = "http://" + addr
	}

	baseURL, err := url.ParseRequestURI(addr)
	if err != nil {
		return errors.Join(ErrAddrParse, err)
	}

	flagAddr = baseURL
	return nil
}
