package config

import (
	"fmt"
	"time"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

const minReportInterval = 1

func getReportConfig(
	flagV, envV *int,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) time.Duration {
	resolve := func(value int, src, name string) time.Duration {
		if value < minReportInterval {
			errStream <- fmt.Errorf(
				"report interval '%d' less '%d', source '%s' name '%s'",
				value, minReportInterval, src, name,
			)
		}
		return time.Duration(value) * time.Second
	}

	if envSet.IsSet(reportEnvName) {
		return resolve(*envV, srcEnv, reportEnvName)
	}
	if flagSet.IsSet(reportFlagName) {
		return resolve(*flagV, srcFlag, "-"+reportFlagName)
	}
	if settings.Report != nil {
		return resolve(*settings.Report, srcSettings, reportSettingsName)
	}
	return time.Duration(reportDefault) * time.Second
}
