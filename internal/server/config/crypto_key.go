package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

func getCryptoKeyFile(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) *os.File {
	resolve := func(p, src, name string) *os.File {
		f, err := os.Open(p)
		if err != nil {
			errStream <- fmt.Errorf(
				"failed to open crypto key path '%s', source '%s' name '%s': %w",
				p, src, name, err,
			)
			return nil
		}
		return f
	}

	if envSet.IsSet(cryptoKeyEnvName) {
		return resolve(*envV, srcEnv, cryptoKeyEnvName)
	}
	if flagSet.IsSet(cryptoKeyFlagName) {
		return resolve(*flagV, srcFlag, cryptoKeyFlagName)
	}
	if settings.CryptoKey != nil {
		return resolve(*settings.CryptoKey, srcSettings, cryptoKeySettingsName)
	}

	errStream <- errors.New("failed to open crypto key file, flag is required")
	return nil
}

func getCryptoKeyData(p *os.File, errStream chan<- error) []byte {
	errText := "failed to read crypto key file"

	if p == nil {
		errStream <- errors.New(errText)
		return nil
	}

	pemData, err := io.ReadAll(p)
	if err != nil {
		errStream <- fmt.Errorf("%s: %w", errText, err)
		return nil
	}
	return pemData
}
