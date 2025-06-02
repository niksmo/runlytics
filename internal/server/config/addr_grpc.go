package config

import (
	"fmt"
	"net"
)

type GRPCAddrConfig struct {
	TCPAddr *net.TCPAddr
}

func NewGRPCAddrConfig(p ConfigParams) (ac GRPCAddrConfig) {
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
	case p.EnvSet.IsSet(grpcAddrEnvName):
		resolveAddr(*p.EnvValues.addr, srcEnv, grpcAddrEnvName)
	case p.FlagSet.IsSet(grpcAddrFlagName):
		resolveAddr(*p.FlagValues.addr, srcFlag, grpcAddrFlagName)
	case p.Settings.Address != nil:
		resolveAddr(*p.Settings.Address, srcSettings, grpcAddrSettingsName)
	default:
		resolveAddr(grpcAddrDefault, "", "")
	}
	return
}
