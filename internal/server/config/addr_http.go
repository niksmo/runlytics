package config

import (
	"fmt"
	"net"
)

type HTTPAddrConfig struct {
	TCPAddr *net.TCPAddr
}

func NewHTTPAddrConfig(p ConfigParams) (ac HTTPAddrConfig) {
	resolveAddr := func(addr, src, name string) {
		TCPAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			p.ErrStream <- fmt.Errorf(
				"invalid TCP address '%s', source '%s' name '%s': %w",
				addr, src, name, err,
			)
			return
		}
		ac.TCPAddr = TCPAddr
	}

	switch {
	case p.EnvSet.IsSet(httpAddrEnvName):
		resolveAddr(*p.EnvValues.addr, srcEnv, httpAddrEnvName)
	case p.FlagSet.IsSet(httpAddrFlagName):
		resolveAddr(*p.FlagValues.addr, srcFlag, httpAddrFlagName)
	case p.Settings.Address != nil:
		resolveAddr(*p.Settings.Address, srcSettings, httpAddrSettingsName)
	default:
		resolveAddr(httpAddrDefault, "", "")
	}
	return
}
