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

	port, err := strconv.Atoi(slice[1])

	if err != nil {
		return errAddr
	}

	a.host = slice[0]
	a.port = port
	return nil
}

var flagAddr *addr = &addr{host: "localhost", port: 8080}

func parseFlags() {
	flag.Var(flagAddr, "a", "Input listening server address, e.g. example.com:8080")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		if err := flagAddr.Set(envAddr); err != nil {
			log.Print(fmt.Errorf("parse env ADDRESS error: %w", err))
		}
	}
}
