package config

import (
	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

func getDSNConfig(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
) string {
	if envSet.IsSet(dsnEnvName) {
		return *envV
	}

	if flagSet.IsSet(dsnFlagName) {
		return *flagV
	}

	if settings.DSN != nil {
		return *settings.DSN
	}

	return ""
}
