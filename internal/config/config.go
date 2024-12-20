package config

import (
	"strconv"
)

type serverConfig struct {
	host string
	port int
}

func New() *serverConfig {
	return &serverConfig{"127.0.0.1", 8080}
}

func (c *serverConfig) Addr() string {
	return c.host + ":" + strconv.Itoa(c.port)
}
