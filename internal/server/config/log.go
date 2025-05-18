package config

import (
	"fmt"
	"strings"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

func getLogConfig(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) string {
	allowed := map[string]struct{}{
		"debug":  {},
		"info":   {},
		"warn":   {},
		"error":  {},
		"dpanic": {},
		"panic":  {},
		"fatal":  {},
	}

	resolve := func(value, src, name string) string {
		_, ok := allowed[value]
		if !ok {
			var s []string
			for lvl := range allowed {
				s = append(s, lvl)
			}
			allowedList := strings.Join(s, ", ")

			errStream <- fmt.Errorf(
				"log level '%s' is not allowed, source '%s' name '%s' allowed values: %s",
				value, src, name, allowedList)

			return ""
		}
		return value
	}

	if envSet.IsSet(logEnvName) {
		return resolve(*envV, srcEnv, logEnvName)
	}

	if flagSet.IsSet(logFlagName) {
		return resolve(*flagV, srcFlag, "-"+logFlagName)
	}

	if settings.Log != nil {
		return resolve(*settings.Log, srcSettings, logSettingsName)
	}

	return logDefault
}
