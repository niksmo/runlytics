package config

import (
	"fmt"
	"net"
)

type TrustedNetConfig struct {
	IPNet *net.IPNet
}

func NewTrustedNetConfig(p ConfigParams) (tc TrustedNetConfig) {
	resolveTrustedNet := func(value, src, name string) {
		_, IPNet, err := net.ParseCIDR(value)
		if err != nil {
			p.ErrStream <- fmt.Errorf(
				"failed to resolve trusted subnet '%s', source '%s' name '%s': %w",
				value, src, name, err,
			)
			return
		}
		tc.IPNet = IPNet
	}

	switch {
	case p.EnvSet.IsSet(trustedNetEnvName):
		resolveTrustedNet(*p.EnvValues.trustedNet, srcEnv, trustedNetEnvName)
	case p.FlagSet.IsSet(trustedNetFlagName):
		resolveTrustedNet(
			*p.FlagValues.trustedNet, srcFlag, "-"+trustedNetFlagName,
		)
	case p.Settings.TrustedNet != nil:
		resolveTrustedNet(
			*p.Settings.TrustedNet, srcSettings, trustedNetSettingsName,
		)
	}
	return
}

func (tc *TrustedNetConfig) IsSet() bool {
	return tc.IPNet != nil
}
