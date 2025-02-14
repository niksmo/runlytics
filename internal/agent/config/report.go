package config

import (
	"os"
	"strconv"
	"time"
)

const (
	reportDefault     = 10
	reportUsage       = "Emitting metrics interval in sec, e.g. 10 (min 1)"
	reportEnv         = "REPORT_INTERVAL"
	minReportInterval = 1
)

func getReportFlag(rawReport int) time.Duration {
	var (
		isEnv    bool
		envValue string
	)
	isCmd := rawReport != reportDefault

	if envValue = os.Getenv(reportEnv); envValue != "" {
		isEnv = true
	}

	if isEnv {
		reportInt, err := strconv.Atoi(envValue)
		if err == nil {
			if reportInt >= minReportInterval {
				return time.Duration(reportInt) * time.Second
			}
			printParamError(
				isEnv, reportEnv, "-p", "should be more or equal 1 second",
			)
		} else {
			printParamError(isEnv, reportEnv, "-p", "invalid report interval")
		}
		isEnv = false
	}

	if isCmd {
		if rawReport >= minReportInterval {
			return time.Duration(rawReport) * time.Second
		}
		printParamError(
			isEnv, reportEnv, "-p", "should be more or equal 1 second",
		)
		isCmd = false
	}

	report := reportDefault * time.Second
	printUsedDefault("report interval", report.String())
	return report
}
