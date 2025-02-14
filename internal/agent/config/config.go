package config

import (
	"flag"
	"fmt"
	"net/url"
	"time"
)

type Config struct {
	logLvl string
	addr   *url.URL
	poll   time.Duration
	report time.Duration
}

func Load() *Config {
	rawLogLvlFlag := flag.String("l", logLvlDefault, logLvlUsage)
	rawAddrFlag := flag.String("a", addrDefault, addrUsage)
	rawPollFlag := flag.Int("p", pollDefault, pollUsage)
	rawReportFlag := flag.Int("r", reportDefault, reportUsage)
	flag.Parse()

	pollFlag := getPollFlag(*rawPollFlag)
	reportFlag := getReportFlag(*rawReportFlag)

	if err := verifyPollVsReport(pollFlag, reportFlag); err != nil {
		panic(err)
	}

	config := Config{
		logLvl: getLogLvlFlag(*rawLogLvlFlag),
		addr:   getAddrFlag(*rawAddrFlag),
		poll:   pollFlag,
		report: reportFlag,
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

func verifyPollVsReport(poll, report time.Duration) error {
	if report >= poll {
		return nil
	}

	return fmt.Errorf("Report should be more or equal poll")
}

func printUsedDefault(configField, value string) {
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
