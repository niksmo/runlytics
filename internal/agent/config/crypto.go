package config

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type CryptoConfig struct {
	Path string
	Data []byte
	f    *os.File
}

func NewCryptoConfig(p ConfigParams) (cc CryptoConfig) {
	cc.initFile(p)
	cc.initData(p.ErrStream)
	return
}

func (cc *CryptoConfig) initFile(p ConfigParams) {
	resolveFile := func(path, src, name string) {
		f, err := os.Open(path)
		if err != nil {
			p.ErrStream <- fmt.Errorf(
				"failed to open cert path '%s', source '%s' name '%s': %w",
				path, src, name, err,
			)
		}
		cc.f = f
	}

	switch {
	case p.EnvSet.IsSet(cryptoKeyEnvName):
		resolveFile(*p.EnvValues.cryptoKey, srcEnv, cryptoKeyEnvName)
	case p.FlagSet.IsSet(cryptoKeyFlagName):
		resolveFile(*p.FlagValues.cryptoKey, srcFlag, cryptoKeyFlagName)
	case p.Settings.CryptoKey != nil:
		resolveFile(*p.Settings.CryptoKey, srcSettings, cryptoKeySettingsName)
	default:
		p.ErrStream <- errors.New("failed to open cert file, flag is required")
	}
}

func (cc *CryptoConfig) initData(errStream chan<- error) {
	if cc.f == nil {
		return
	}

	pemData, err := io.ReadAll(cc.f)
	if err != nil {
		errStream <- fmt.Errorf("%s: %w", "failed to read cert file", err)
		return
	}
	cc.Data = pemData
}
