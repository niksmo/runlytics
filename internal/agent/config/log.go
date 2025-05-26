package config

import (
	"fmt"
	"strings"
)

var (
	allowedLogLvl = map[string]struct{}{
		"debug":  {},
		"info":   {},
		"warn":   {},
		"error":  {},
		"dpanic": {},
		"panic":  {},
		"fatal":  {},
	}
)

type LogConfig struct {
	Level string
}

func NewLogConfig(p ConfigParams) (lc LogConfig) {
	resolveLogLevel := func(value, src, name string) string {
		_, ok := allowedLogLvl[value]
		if !ok {

			p.ErrStream <- fmt.Errorf(
				"log level '%s' is not allowed, source '%s' name '%s' allowed values: %s",
				value, src, name, allowedLogStr())

			return ""
		}
		return value
	}

	switch {
	case p.EnvSet.IsSet(logEnvName):
		lc.Level = resolveLogLevel(*p.EnvValues.log, srcEnv, logEnvName)
	case p.FlagSet.IsSet(logFlagName):
		lc.Level = resolveLogLevel(*p.FlagValues.log, srcFlag, "-"+logFlagName)
	case p.Settings.Log != nil:
		lc.Level = resolveLogLevel(*p.Settings.Log, srcSettings, logSettingsName)
	default:
		lc.Level = logDefault
	}
	return
}

func allowedLogStr() string {
	var s []string
	for lvl := range allowedLogLvl {
		s = append(s, lvl)
	}
	return strings.Join(s, ", ")
}
