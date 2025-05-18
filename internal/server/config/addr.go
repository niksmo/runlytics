package config

import (
	"fmt"
	"net"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

func getAddrConfig(
	flagV, envV *string,
	flagSet *flag.FlagSet,
	envSet *env.EnvSet,
	settings settings,
	errStream chan<- error,
) *net.TCPAddr {

	resolveAddr := func(addr, src, name string) *net.TCPAddr {
		TCPAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			errStream <- fmt.Errorf(
				"address '%s' is not valid TCP address, source '%s' name '%s': %w",
				addr, src, name, err,
			)
			return nil
		}
		return TCPAddr

	}

	if envSet.IsSet(addrEnvName) {
		return resolveAddr(*envV, srcEnv, addrEnvName)
	}
	if flagSet.IsSet(addrFlagName) {
		return resolveAddr(*flagV, srcFlag, "-"+addrFlagName)
	}
	if settings.Address != nil {
		return resolveAddr(*settings.Address, srcSettings, addrSettingsName)
	}

	TCPAddr, _ := net.ResolveTCPAddr("tcp", addrDefault)
	return TCPAddr
}
