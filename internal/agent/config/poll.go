package config

import (
	"os"
	"strconv"
	"time"
)

const (
	pollDefault     = 2
	pollUsage       = "Polling collecting metrics interval in sec, e.g. 5 (min 1)"
	pollEnv         = "POLL_INTERVAL"
	minPollInterval = 1
)

func getPollFlag(rawPoll int) time.Duration {
	var (
		isEnv    bool
		envValue string
	)
	isCmd := rawPoll != pollDefault

	if envValue = os.Getenv(pollEnv); envValue != "" {
		isEnv = true
	}

	if isEnv {
		pollInt, err := strconv.Atoi(envValue)
		if err == nil {
			if pollInt >= minPollInterval {
				return time.Duration(pollInt) * time.Second
			}
			printParamError(
				isEnv, pollEnv, "-p", "should be more or equal 1 second",
			)
		} else {
			printParamError(isEnv, pollEnv, "-p", "invalid poll interval")
		}
		isEnv = false
	}

	if isCmd {
		if rawPoll >= minPollInterval {
			return time.Duration(rawPoll) * time.Second
		}
		printParamError(
			isEnv, pollEnv, "-p", "should be more or equal 1 second",
		)
		isCmd = false
	}

	poll := pollDefault * time.Second
	printUsedDefault("poll interval", poll.String())
	return poll
}
