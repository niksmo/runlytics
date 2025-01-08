package main

import (
	"errors"
	"flag"
	"net/url"
	"time"
)

var (
	flagAddr   *addr = &addr{&url.URL{Host: "localhost:8080", Scheme: "http"}}
	flagPoll   time.Duration
	flagReport time.Duration
)

var ()

type addr struct {
	url *url.URL
}

func (a *addr) URL() *url.URL {
	return a.url.ResolveReference(a.url)
}

func (a *addr) String() string {
	return a.url.String()
}

func (a *addr) Set(v string) error {
	baseURL, err := url.ParseRequestURI(v)
	if err != nil {
		return err
	}

	if !(baseURL.Scheme == "http" || baseURL.Scheme == "https") {
		return errors.New("supports only http or https scheme")
	}

	if baseURL.Host == "" {
		return errors.New("pass absolute URL, e.g. http://one.two.com/path1/id")
	}

	a.url = baseURL

	return nil
}

func parseFlags() {
	flag.Var(
		flagAddr,
		"a",
		"Host address for metrics emitting, example: example.com:8080",
	)

	flag.DurationVar(&flagPoll,
		"p",
		time.Duration(2*time.Second),
		"Polling collecting metrics interval in sec, example: 5s",
	)

	flag.DurationVar(&flagReport,
		"r",
		time.Duration(10*time.Second),
		"Emiting metrics interval in sec, example: 10s",
	)

	flag.Parse()
}
