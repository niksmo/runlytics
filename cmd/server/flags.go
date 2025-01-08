package main

import (
	"errors"
	"flag"
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

	a.host = slice[0]

	port, err := strconv.Atoi(slice[1])

	if err != nil {
		return errAddr
	}

	a.port = port
	return nil
}

var flagAddr *addr = &addr{host: "localhost", port: 8080}

func parseFlags() {
	flag.Var(flagAddr, "a", "Input listening server address, example: example.com:8080")
	flag.Parse()
}
