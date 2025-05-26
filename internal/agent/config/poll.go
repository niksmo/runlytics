package config

import (
	"fmt"
	"time"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

const minPollInterval = 1

func getPollConfig(
	flagV, envV *int,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) time.Duration {
	resolve := func(value int, src, name string) time.Duration {
		if value < minPollInterval {
			errStream <- fmt.Errorf(
				"poll interval '%d' less '%d', source '%s' name '%s'",
				value, minPollInterval, src, name,
			)
		}
		return time.Duration(value) * time.Second
	}

	if envSet.IsSet(pollEnvName) {
		return resolve(*envV, srcEnv, pollEnvName)
	}
	if flagSet.IsSet(pollFlagName) {
		return resolve(*flagV, srcFlag, "-"+pollFlagName)
	}
	if settings.Poll != nil {
		return resolve(*settings.Poll, srcSettings, pollSettingsName)
	}
	return time.Duration(pollDefault) * time.Second
}
