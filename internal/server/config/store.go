package config

import (
	"fmt"
	"os"
	"time"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

func getStoreFileConfig(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) *os.File {
	resolve := func(p, src, name string) *os.File {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			errStream <- fmt.Errorf(
				"failed to open store path '%s', source '%s' name '%s': %w",
				p, src, name, err,
			)
			return nil
		}
		return f
	}

	if envSet.IsSet(storeEnvName) {
		return resolve(*envV, srcEnv, storeEnvName)
	}

	if flagSet.IsSet(storeFlagName) {
		return resolve(*flagV, srcFlag, "-"+storeFlagName)
	}

	if settings.Restore != nil {
		return resolve(*settings.StoreFile, srcSettings, storeSettingsName)
	}

	f, _ := os.OpenFile(storeDefaultPath, os.O_CREATE|os.O_RDWR, 0644)
	return f
}

func getStoreIntervalConfig(
	flagV, envV *int,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) time.Duration {
	resolve := func(value int, src, name string) time.Duration {
		if value < 0 {
			errStream <- fmt.Errorf(
				"store interval '%d' less zero, source '%s' name '%s'",
				value, src, name,
			)
			return 0
		}

		return time.Second * time.Duration(value)
	}

	if envSet.IsSet(storeIntervalEnvName) {
		return resolve(*envV, srcEnv, storeIntervalEnvName)
	}

	if flagSet.IsSet(storeIntervalFlagName) {
		return resolve(*flagV, srcFlag, "-"+storeIntervalFlagName)
	}

	if settings.StoreInterval != nil {
		return resolve(*settings.StoreInterval, srcSettings, storeIntervalSettingsName)
	}

	return time.Second * time.Duration(storeIntervalDefault)
}

func getStoreRestoreConfig(
	flagV, envV *bool,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
) bool {

	if envSet.IsSet(storeRestoreEnvName) {
		return *envV
	}

	if flagSet.IsSet(storeRestoreFlagName) {
		return *flagV
	}

	if settings.Restore != nil {
		return *settings.Restore
	}

	return storeRestoreDefault
}
