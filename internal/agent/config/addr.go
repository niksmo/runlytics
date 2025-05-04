package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

const (
	addrDefault = "http://localhost:8080"
	addrUsage   = "Host address for metrics emitting, e.g. http://example.com:8080"
	addrEnv     = "ADDRESS"
)

var allowedScheme = map[string]struct{}{
	"http":  {},
	"https": {},
}

func getAddrFlag(addr string) *url.URL {
	var (
		isEnv    bool
		envValue string
	)
	cmdValue := addr
	isCmd := cmdValue != addrDefault

	if envValue = os.Getenv(addrEnv); envValue != "" {
		isEnv = true
	}

	if isEnv {
		URL, err := parseAddrToURL(envValue)
		if err == nil {
			return URL
		}
		printParamError(isEnv, addrEnv, "-a", "invalid addr")
		isEnv = false
	}

	if isCmd {
		URL, err := parseAddrToURL(cmdValue)
		if err == nil {
			return URL
		}
		printParamError(isEnv, addrEnv, "-a", "invalid addr")
	}

	URL, _ := url.ParseRequestURI(addrDefault)

	return URL
}

func parseAddrToURL(addr string) (*url.URL, error) {
	var hasScheme bool
	for scheme := range allowedScheme {
		if strings.HasPrefix(addr, scheme) {
			hasScheme = true
			break
		}
	}
	if !hasScheme {
		addr = "http://" + addr
	}

	URL, err := url.ParseRequestURI(addr)
	if err != nil || URL.Host == "" {
		return nil, errors.New("invalid addr: " + addr)
	}

	return URL, nil

}
