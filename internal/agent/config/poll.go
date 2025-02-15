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
	printMinIntervalErr := func(isEnv bool) {
		printParamError(
			isEnv, pollEnv, "-p", "should be more or equal 1 second",
		)
	}

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
		switch {
		case err != nil:
			printParamError(isEnv, pollEnv, "-p", "invalid poll interval")
		case pollInt >= minPollInterval:
			return time.Duration(pollInt) * time.Second
		default:
			printMinIntervalErr(isEnv)
		}
		isEnv = false
	}

	if isCmd {
		if rawPoll >= minPollInterval {
			return time.Duration(rawPoll) * time.Second
		}
		printMinIntervalErr(isEnv)
		isCmd = false
	}

	poll := pollDefault * time.Second
	printUsedDefault("poll interval", poll.String())
	return poll
}
