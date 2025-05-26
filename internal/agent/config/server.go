package config

import (
	"fmt"
	"net"
	"net/url"
)

type ServerConfig struct {
	Addr         *net.TCPAddr
	Scheme, Path string
}

func NewServerConfig(p ConfigParams) (sc ServerConfig) {
	resolveAddr := func(addr, src, name string) *net.TCPAddr {
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			p.ErrStream <- fmt.Errorf(
				"invalid address '%s', source '%s' name '%s': %w",
				addr, src, name, err,
			)
		}
		return tcpAddr
	}

	switch {
	case p.EnvSet.IsSet(addrEnvName):
		sc.Addr = resolveAddr(*p.EnvValues.addr, srcEnv, addrEnvName)
	case p.FlagSet.IsSet(addrFlagName):
		sc.Addr = resolveAddr(*p.FlagValues.addr, srcFlag, "-"+addrFlagName)
	case p.Settings.Address != nil:
		sc.Addr = resolveAddr(
			*p.Settings.Address, srcSettings, addrSettingsName,
		)
	default:
		sc.Addr = resolveAddr(addrDefault, "", "")
	}

	sc.Scheme = "http"
	sc.Path = "updates"

	return
}

func (sc *ServerConfig) URL() string {
	URL := new(url.URL)
	URL.Scheme = sc.Scheme
	URL.Host = sc.Addr.String()
	URL.Path = sc.Path
	return URL.String()
}
