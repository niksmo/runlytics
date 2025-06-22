package config

import (
	"fmt"
	"net"
	"net/url"
)

const (
	scheme = "http"
	path   = "updates"
)

type ServerConfig struct {
	HTTPAddr, GRPCAddr *net.TCPAddr
	Scheme, Path       string
}

func NewServerConfig(p ConfigParams) (sc ServerConfig) {
	sc.Scheme = scheme
	sc.Path = path
	sc.initHTTPAddr(p)
	sc.initGRPCAddr(p)
	return
}

func (sc *ServerConfig) initHTTPAddr(p ConfigParams) {
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
		sc.HTTPAddr = resolveAddr(*p.EnvValues.addr, srcEnv, addrEnvName)
	case p.FlagSet.IsSet(addrFlagName):
		sc.HTTPAddr = resolveAddr(*p.FlagValues.addr, srcFlag, "-"+addrFlagName)
	case p.Settings.Address != nil:
		sc.HTTPAddr = resolveAddr(
			*p.Settings.Address, srcSettings, addrSettingsName,
		)
	default:
		sc.HTTPAddr = resolveAddr(addrDefault, "", "")
	}
}

func (sc *ServerConfig) URL() string {
	URL := new(url.URL)
	URL.Scheme = sc.Scheme
	URL.Host = sc.HTTPAddr.String()
	URL.Path = sc.Path
	return URL.String()
}

func (sc *ServerConfig) initGRPCAddr(p ConfigParams) {
	resolveGPRCAddr := func(addr, src, name string) *net.TCPAddr {
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
	case p.EnvSet.IsSet(grpcEnvName):
		sc.GRPCAddr = resolveGPRCAddr(*p.EnvValues.addr, srcEnv, grpcEnvName)
	case p.FlagSet.IsSet(grpcFlagName):
		sc.GRPCAddr = resolveGPRCAddr(*p.FlagValues.addr, srcFlag, "-"+grpcFlagName)
	case p.Settings.GRPC != nil:
		sc.GRPCAddr = resolveGPRCAddr(
			*p.Settings.GRPC, srcSettings, grpcSettingsName,
		)
	default:
		sc.GRPCAddr = resolveGPRCAddr(grpcDefault, "", "")
	}
}
