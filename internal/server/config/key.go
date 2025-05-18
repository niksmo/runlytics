package config

import (
	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

func getHashKeyConfig(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
) string {
	if envSet.IsSet(hashKeyEnvName) {
		return *envV
	}

	if flagSet.IsSet(hashKeyFlagName) {
		return *flagV
	}

	if settings.HashKey != nil {
		return *settings.HashKey
	}

	return ""
}
