package config

import (
	"os"
	"runtime"
	"strconv"
)

const (
	minRateLimit   = 1
	rateLimitUsage = "Emitting rate limit, e.g. 8 (min 1)"
	rateLimitEnv   = "RATE_LIMIT"
)

var rateLimitDefault = runtime.NumCPU()

func getRateLimitFlag(rawRate int) int {
	var (
		isEnv    bool
		envValue string
	)

	printMinRateErr := func(isEnv bool) {
		printParamError(
			isEnv, rateLimitEnv, "-l", "should be more or equal 1",
		)
	}

	isCmd := rawRate != rateLimitDefault

	if envValue = os.Getenv(rateLimitEnv); envValue != "" {
		isEnv = true
	}

	if isEnv {
		rateInt, err := strconv.Atoi(envValue)
		switch {
		case err != nil:
			printParamError(isEnv, reportEnv, "-l", "invalid rate limit")
		case rateInt >= minRateLimit:
			return rateInt
		default:
			printMinRateErr(isEnv)
		}
		isEnv = false
	}

	if isCmd {
		if rawRate >= minRateLimit {
			return rawRate
		}
		printMinRateErr(isEnv)
	}

	printUsedDefault("rate limit", rateLimitDefault)
	return rateLimitDefault
}
