package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

var allowedScheme = map[string]struct{}{
	"http":  {},
	"https": {},
}

func getAddrConfig(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) *url.URL {

	resolve := func(value, src, name string) *url.URL {
		URL, err := parseAddrToURL(value)
		if err != nil {
			errStream <- fmt.Errorf(
				"invalid address '%s', source '%s' name '%s': %w",
				value, src, name, err,
			)
		}
		return URL
	}

	if envSet.IsSet(addrEnvName) {
		return resolve(*envV, srcEnv, addrEnvName)
	}
	if flagSet.IsSet(addrFlagName) {
		return resolve(*flagV, srcFlag, "-"+addrFlagName)
	}
	if settings.Address != nil {
		return resolve(*settings.Address, srcSettings, addrSettingsName)
	}

	URL, _ := parseAddrToURL(addrDefault)
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
		return nil, err
	}

	return URL, nil

}
