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

	printMinIntervalErr := func(isEnv bool) {
		printParamError(
			isEnv, pollEnv, "-p", "should be more or equal 1 second",
		)
	}

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
		switch {
		case err != nil:
			printParamError(isEnv, reportEnv, "-p", "invalid report interval")
		case reportInt >= minReportInterval:
			return time.Duration(reportInt) * time.Second
		default:
			printMinIntervalErr(isEnv)
		}
		isEnv = false
	}

	if isCmd {
		if rawReport >= minReportInterval {
			return time.Duration(rawReport) * time.Second
		}
		printMinIntervalErr(isEnv)
		isCmd = false
	}

	report := reportDefault * time.Second
	printUsedDefault("report interval", report.String())
	return report
}
