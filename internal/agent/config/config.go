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
	addr      *url.URL
	key       string
	cryptoKey cryptoKey
	logLvl    string
	poll      time.Duration
	rateLimit int
	report    time.Duration
}

func Load() *Config {
	rawLogLvlFlag := flag.String("log", logLvlDefault, logLvlUsage)
	rawAddrFlag := flag.String("a", addrDefault, addrUsage)
	rawPollFlag := flag.Int("p", pollDefault, pollUsage)
	rawReportFlag := flag.Int("r", reportDefault, reportUsage)
	rawKeyFlag := flag.String("k", keyDefault, keyUsage)
	rawRateLimitFlag := flag.Int("l", rateLimitDefault, rateLimitUsage)

	rawCryptoKeyFlag := flag.String(
		"crypto-key", cryptoKeyDefault, cryptoKeyUsage,
	)
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
		cryptoKey: getCryptoKeyFlag(*rawCryptoKeyFlag),
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

// CryptoKeyPath returns cert path.
func (c *Config) CryptoKeyPath() string {
	return c.cryptoKey.path
}

// CryptoKeyData returns cert file data.
func (c *Config) CryptoKeyData() []byte {
	return c.cryptoKey.pemData
}

func verifyPollVsReport(poll, report time.Duration) error {
	if report >= poll {
		return nil
	}

	return fmt.Errorf("Report should be more or equal poll")
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
