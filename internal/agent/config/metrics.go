package config

import (
	"fmt"
	"time"
)

const (
	minPollInterval   = 1
	minReportInterval = 1
	minRateLimit      = 1
)

type MetricsConfig struct {
	Poll, Report time.Duration
	RateLimit    int
	JobsBuf      int
	JobsErrBuf   int
}

func NewMetricsConfig(p ConfigParams) (mc MetricsConfig) {
	mc.initPoll(p)
	mc.initReport(p)
	mc.initRateLimit(p)
	mc.verifyPollVsReport(p.ErrStream)
	return
}

func (mc *MetricsConfig) initPoll(p ConfigParams) {
	resolvePoll := func(value int, src, name string) {
		if value < minPollInterval {
			p.ErrStream <- fmt.Errorf(
				"poll interval '%d' less '%d', source '%s' name '%s'",
				value, minPollInterval, src, name,
			)
		}
		mc.Poll = time.Duration(value) * time.Second
	}

	switch {
	case p.EnvSet.IsSet(pollEnvName):
		resolvePoll(*p.EnvValues.poll, srcEnv, pollEnvName)
	case p.FlagSet.IsSet(pollFlagName):
		resolvePoll(*p.FlagValues.poll, srcFlag, pollFlagName)
	case p.Settings.Poll != nil:
		resolvePoll(*p.Settings.Poll, srcSettings, pollSettingsName)
	default:
		mc.Poll = time.Duration(pollDefault) * time.Second
	}
}

func (mc *MetricsConfig) initReport(p ConfigParams) {
	resolveReport := func(value int, src, name string) {
		if value < minReportInterval {
			p.ErrStream <- fmt.Errorf(
				"report interval '%d' less '%d', source '%s' name '%s'",
				value, minReportInterval, src, name,
			)
		}
		mc.Report = time.Duration(value) * time.Second
	}

	switch {
	case p.EnvSet.IsSet(reportEnvName):
		resolveReport(*p.EnvValues.report, srcEnv, reportEnvName)
	case p.FlagSet.IsSet(reportFlagName):
		resolveReport(*p.FlagValues.report, srcFlag, reportFlagName)
	case p.Settings.Report != nil:
		resolveReport(*p.Settings.Report, srcSettings, reportSettingsName)
	default:
		mc.Report = time.Duration(reportDefault) * time.Second
	}
}

func (mc *MetricsConfig) initRateLimit(p ConfigParams) {
	resolveRateLimit := func(value int, src, name string) {
		if value < minRateLimit {
			p.ErrStream <- fmt.Errorf(
				"rate limit '%d' less '%d', source '%s' name '%s'",
				value, minRateLimit, src, name,
			)
		}
		mc.RateLimit = value
	}

	switch {
	case p.EnvSet.IsSet(rateLimitEnvName):
		resolveRateLimit(*p.EnvValues.rateLimit, srcEnv, rateLimitEnvName)
	case p.FlagSet.IsSet(reportFlagName):
		resolveRateLimit(*p.FlagValues.rateLimit, srcFlag, reportFlagName)
	case p.Settings.RateLimit != nil:
		resolveRateLimit(*p.Settings.RateLimit, srcSettings, rateLimitSettingsName)
	default:
		mc.RateLimit = rateLimitDefault
	}

}

func (mc *MetricsConfig) verifyPollVsReport(errStream chan<- error) {
	if mc.Report < mc.Poll {
		errStream <- fmt.Errorf(
			"report '%v' should be more or equal poll '%v'",
			mc.Report.Seconds(), mc.Poll.Seconds(),
		)
	}
}
