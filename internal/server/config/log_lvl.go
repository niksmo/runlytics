package config

import (
	"os"
	"strings"
)

const (
	logLvlDefault = "info"
	logLvlUsage   = "Logging level, e.g. 'debug'"
	logLvlEnv     = "LOG_LVL"
)

func getLogLvlFlag(logLvl string) string {
	allowed := map[string]struct{}{
		"debug":  {},
		"info":   {},
		"warn":   {},
		"error":  {},
		"dpanic": {},
		"panic":  {},
		"fatal":  {},
	}

	isEnv := false
	if envValue := os.Getenv(logLvlEnv); envValue != "" {
		isEnv = true
		logLvl = envValue
	}

	if _, ok := allowed[strings.ToLower(logLvl)]; !ok {
		text := "error: level is not allowed"
		if isEnv {
			printEnvParamError(logLvlEnv, text)
		} else {
			printCliParamError("-l", text)
		}
		logLvl = logLvlDefault
		printUsedDefault("logging level", logLvlDefault)
	}

	return logLvl
}
