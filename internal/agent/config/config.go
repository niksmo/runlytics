package config

import (
	"flag"
	"fmt"
	"net/url"
	"time"
)

const jobsBuf = 1024
const jobsErrBuf = 128

type Config struct {
	logLvl    string
	addr      *url.URL
	poll      time.Duration
	report    time.Duration
	key       string
	rateLimit int
}

func Load() *Config {
	rawLogLvlFlag := flag.String("log", logLvlDefault, logLvlUsage)
	rawAddrFlag := flag.String("a", addrDefault, addrUsage)
	rawPollFlag := flag.Int("p", pollDefault, pollUsage)
	rawReportFlag := flag.Int("r", reportDefault, reportUsage)
	rawKeyFlag := flag.String("k", keyDefault, keyUsage)
	rawRateLimitFlag := flag.Int("l", rateLimitDefault, rateLimitUsage)
	flag.Parse()

	pollFlag := getPollFlag(*rawPollFlag)
	reportFlag := getReportFlag(*rawReportFlag)

	if err := verifyPollVsReport(pollFlag, reportFlag); err != nil {
		panic(err)
	}

	config := Config{
		logLvl:    getLogLvlFlag(*rawLogLvlFlag),
		addr:      getAddrFlag(*rawAddrFlag),
		poll:      pollFlag,
		report:    reportFlag,
		key:       getKeyFlag(*rawKeyFlag),
		rateLimit: getRateLimitFlag(*rawRateLimitFlag),
	}
	return &config
}

func (c *Config) LogLvl() string {
	return c.logLvl
}

func (c *Config) Addr() *url.URL {
	URL := *c.addr
	return &URL
}

func (c *Config) Poll() time.Duration {
	return c.poll
}

func (c *Config) Report() time.Duration {
	return c.report
}

func (c *Config) Key() string {
	return c.key
}

func (c *Config) RateLimit() int {
	return c.rateLimit
}

func (c *Config) JobsBuf() int {
	return jobsBuf
}

func (c *Config) JobsErrBuf() int {
	return jobsErrBuf
}

func (c *Config) HTTPClientTimeout() time.Duration {
	return c.Report() - 100*time.Millisecond
}

func verifyPollVsReport(poll, report time.Duration) error {
	if report >= poll {
		return nil
	}

	return fmt.Errorf("Report should be more or equal poll")
}

func printUsedDefault(configField string, value any) {
	fmt.Println("Used default", configField+":", value)
}

func printParamError(isEnv bool, envP, cmdP, errText string) {
	var prefix string
	var p string
	if isEnv {
		prefix = "Env param"
		p = envP
	} else {
		prefix = "Cmd param"
		p = cmdP
	}
	fmt.Println(prefix, p, errText)
}
