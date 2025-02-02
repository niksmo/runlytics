package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	saveIntervalDefault = 300
	saveIntervalUsage   = "File storage save interval, '0' is sync"
	saveIntervalEnv     = "STORE_INTERVAL"
)

func getSaveIntervalFlag(interval int) time.Duration {
	isEnv := false
	if envValue := os.Getenv(saveIntervalEnv); envValue != "" {
		isEnv = true
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			printEnvParamError(saveIntervalEnv, "error: value should be integer")
			interval = saveIntervalDefault
			printUsedDefault(
				"file storage save interval",
				fmt.Sprintf("%v", saveIntervalDefault),
			)
		} else {
			interval = intValue
		}
	}

	if interval < 0 {
		text := "error: should more or equal '0'"
		if isEnv {
			printEnvParamError(saveIntervalEnv, text)
		} else {
			printCliParamError("-i", text)
		}
	}

	return time.Duration(interval) * time.Second
}
