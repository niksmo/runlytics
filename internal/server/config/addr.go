package config

import (
	"fmt"
	"net"
)

type AddrConfig struct {
	TCPAddr *net.TCPAddr
}

func NewAddrConfig(p ConfigParams) (ac AddrConfig) {
	resolveAddr := func(addr, src, name string) {
		TCPAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			p.ErrStream <- fmt.Errorf(
				"address '%s' is not valid TCP address, source '%s' name '%s': %w",
				addr, src, name, err,
			)
			return
		}
		ac.TCPAddr = TCPAddr
	}

	switch {
	case p.EnvSet.IsSet(addrEnvName):
		resolveAddr(*p.EnvValues.addr, srcEnv, addrEnvName)
	case p.FlagSet.IsSet(addrFlagName):
		resolveAddr(*p.FlagValues.addr, srcFlag, addrFlagName)
	case p.Settings.Address != nil:
		resolveAddr(*p.Settings.Address, srcSettings, addrSettingsName)
	default:
		resolveAddr(addrDefault, "", "")
	}
	return
}
