package config

import (
	"net"
	"os"
)

const (
	addrDefault = "localhost:8080"
	addrEnv     = "ADDRESS"
	addrUsage   = "Listening server address, e.g. '127.0.0.1:8080'"
)

func getAddrFlag(addr string) *net.TCPAddr {
	printError := func(isEnv bool, text string) {
		printParamError(isEnv, addrEnv, "-a", text)
	}

	isEnv := false
	if envValue := os.Getenv(addrEnv); envValue != "" {
		isEnv = true
		addr = envValue
	}

	TCPAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		text := "error: " + err.Error()
		printError(isEnv, text)
		TCPAddr, _ = net.ResolveTCPAddr("tcp", addrDefault)
		printUsedDefault("address", addrDefault)
	}

	return TCPAddr
}
