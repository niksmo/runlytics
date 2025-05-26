package config

import (
	"fmt"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

const minRateLimit = 1

func getRateLimitConfig(
	flagV, envV *int,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) int {
	resolve := func(value int, src, name string) int {
		if value < minRateLimit {
			errStream <- fmt.Errorf(
				"rate limit '%d' less '%d', source '%s' name '%s'",
				value, minPollInterval, src, name,
			)
		}
		return value
	}

	if envSet.IsSet(rateLimitEnvName) {
		return resolve(*envV, srcEnv, rateLimitEnvName)
	}
	if flagSet.IsSet(rateLimitFlagName) {
		return resolve(*flagV, srcFlag, "-"+rateLimitFlagName)
	}
	if settings.RateLimit != nil {
		return resolve(*settings.RateLimit, srcSettings, rateLimitSettingsName)
	}
	return rateLimitDefault
}
