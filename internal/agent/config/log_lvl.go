package config

import (
	"os"
	"strings"
)

const (
	logLvlDefault = "info"
	logLvlUsage   = `Logging level, e.g. "debug"`
	logLvlEnv     = "LOG_LVL"
)

func getLogLvlFlag(logLvl string) string {
	printError := func(isEnv bool, text string) {
		printParamError(isEnv, logLvlEnv, "-log", text)
	}

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
		printError(isEnv, "error: level is not allowed")
		logLvl = logLvlDefault
	}

	return logLvl
}
