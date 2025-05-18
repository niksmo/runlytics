package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/failprint"
	"github.com/niksmo/runlytics/pkg/flag"
)

const (
	srcEnv  = "env"
	srcFlag = "arg"
)

var srcSettings = "settings.json" // change dynamically

const (
	addrFlagName     = "a"
	addrEnvName      = "ADDRESS"
	addrSettingsName = "address"
	addrDefault      = "http://localhost:8080"
	addrUsage        = "Host address for metrics emitting, e.g. 'http://example.com:8080'"

	logFlagName     = "log"
	logEnvName      = "LOG_LVL"
	logSettingsName = "log"
	logDefault      = "info"
	logUsage        = "Logging level, e.g. 'debug'"

	pollFlagName     = "p"
	pollEnvName      = "POLL_INTERVAL"
	pollSettingsName = "poll_interval"
	pollDefault      = 2
	pollUsage        = "Collecting metrics interval in sec, e.g. '5' (min '1')"

	reportFlagName     = "r"
	reportEnvName      = "REPORT_INTERVAL"
	reportSettingsName = "report_interval"
	reportDefault      = 10
	reportUsage        = "Emitting metrics interval in sec, e.g. '10' (min '1')"

	hashKeyFlagName     = "k"
	hashKeyEnvName      = "KEY"
	hashKeySettingsName = "hash_key"
	hashKeyDefault      = ""
	hashKeyUsage        = "Sercret key for hashing." +
		" Set HashSHA256 value in HTTP header requests (optional)"

	rateLimitFlagName     = "l"
	rateLimitEnvName      = "RATE_LIMIT"
	rateLimitSettingsName = "rate_limit"
	rateLimitUsage        = "Emitting rate limit, e.g. 8 (min 1)"

	cryptoKeyFlagName     = "crypto-key"
	cryptoKeyEnvName      = "CRYPTO_KEY"
	cryptoKeySettingsName = "crypto_key"
	cryptoKeyDefault      = ""
	cryptoKeyUsage        = "Cert path, e.g. '/folder/cert.pem' (required)"

	configFileFlagName = "config"
	configFileEnvName  = "CONFIG"
	configFileDefault  = ""
	configFileUsage    = "Path to json config file, e.g. '/folder/to/config.json' (optional)"
)

var rateLimitDefault = runtime.NumCPU()

const (
	jobsBuf    = 1024
	jobsErrBuf = 128
)

type flagValues struct {
	addr       *string
	log        *string
	poll       *int
	report     *int
	hashKey    *string
	rateLimit  *int
	cryptoKey  *string
	configFile *string
}

type envValues struct {
	addr       *string
	log        *string
	poll       *int
	report     *int
	hashKey    *string
	rateLimit  *int
	cryptoKey  *string
	configFile *string
}

type settings struct {
	Address   *string `json:"address"`
	Log       *string `json:"log"`
	Poll      *int    `json:"poll_interval"`
	Report    *int    `json:"report_interval"`
	HashKey   *string `json:"hash_key"`
	RateLimit *int    `json:"rate_limit"`
	CryptoKey *string `json:"crypto_key"`
}

func newSettings(path string) (settings, error) {
	f, err := os.Open(path)
	if err != nil {
		return settings{}, fmt.Errorf("failed to read settings file: %w", err)
	}

	var s settings
	err = json.NewDecoder(f).Decode(&s)
	if err != nil {
		return s, fmt.Errorf("failed to decode settings file: %w", err)
	}
	return s, nil
}

type Config struct {
	addr             *url.URL
	logLvl           string
	poll             time.Duration
	report           time.Duration
	hashKey          string
	rateLimit        int
	cryptoKeyFile    *os.File
	cryptoKeyPEMData []byte
}

func Load() *Config {
	var (
		flagSet       = flag.New()
		envSet        = env.New()
		cryptoKeyData []byte
	)
	flagV := setupFlagValues(flagSet)
	envV := setupEnvValues(envSet)

	flagSet.Parse()
	if err := envSet.Parse(); err != nil {
		failprint.Println(err)
	}

	settings := loadSettings(flagV.configFile, envV.configFile, flagSet, envSet)

	errStream := make(chan error)
	go failprint.PrintFailWorker(errStream, failprint.ExitOnError)

	addrConfig := getAddrConfig(
		flagV.addr, envV.addr, flagSet, envSet, settings, errStream,
	)

	logConfig := getLogConfig(
		flagV.log, envV.log, flagSet, envSet, settings, errStream,
	)

	pollConfig := getPollConfig(
		flagV.poll, envV.poll, flagSet, envSet, settings, errStream,
	)

	reportConfig := getReportConfig(
		flagV.report, envV.report, flagSet, envSet, settings, errStream,
	)

	verifyPollVsReport(pollConfig, reportConfig, errStream)

	hashKeyConfig := getHashKeyConfig(
		flagV.hashKey, envV.hashKey, flagSet, envSet, settings,
	)

	rateLimitConfig := getRateLimitConfig(
		flagV.rateLimit, envV.rateLimit, flagSet, envSet, settings, errStream,
	)

	cryptoKeyFile := getCryptoKeyFile(
		flagV.cryptoKey, envV.cryptoKey, flagSet, envSet, settings, errStream,
	)

	if cryptoKeyFile != nil {
		cryptoKeyData = getCryptoKeyData(cryptoKeyFile, errStream)
	}

	close(errStream)

	return &Config{
		addr:             addrConfig,
		logLvl:           logConfig,
		poll:             pollConfig,
		report:           reportConfig,
		hashKey:          hashKeyConfig,
		rateLimit:        rateLimitConfig,
		cryptoKeyFile:    cryptoKeyFile,
		cryptoKeyPEMData: cryptoKeyData,
	}

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
	return c.hashKey
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
	if c.cryptoKeyFile != nil {
		return c.cryptoKeyFile.Name()
	}
	return ""
}

// CryptoKeyData returns cert file data.
func (c *Config) CryptoKeyData() []byte {
	return c.cryptoKeyPEMData
}

func verifyPollVsReport(
	poll, report time.Duration, errStream chan<- error,
) {
	if report < poll {
		errStream <- fmt.Errorf(
			"report '%v' should be more or equal poll '%v'",
			report.Seconds(), poll.Seconds(),
		)
	}
}

func setupFlagValues(flagSet *flag.FlagSet) flagValues {
	var fv flagValues

	fv.addr = flagSet.String(addrFlagName, addrDefault, addrUsage)
	fv.log = flagSet.String(logFlagName, logDefault, logUsage)
	fv.poll = flagSet.Int(pollFlagName, pollDefault, pollUsage)
	fv.report = flagSet.Int(reportFlagName, reportDefault, reportUsage)
	fv.hashKey = flagSet.String(hashKeyFlagName, hashKeyDefault, hashKeyUsage)

	fv.rateLimit = flagSet.Int(
		rateLimitFlagName, rateLimitDefault, rateLimitUsage,
	)
	fv.cryptoKey = flagSet.String(
		cryptoKeyFlagName, cryptoKeyDefault, cryptoKeyUsage,
	)
	fv.configFile = flagSet.String(
		configFileFlagName, configFileDefault, configFileUsage,
	)

	return fv
}

func setupEnvValues(envSet *env.EnvSet) envValues {
	var ev envValues
	ev.addr = envSet.String(addrEnvName)
	ev.log = envSet.String(logEnvName)
	ev.poll = envSet.Int(pollEnvName)
	ev.report = envSet.Int(reportEnvName)
	ev.hashKey = envSet.String(hashKeyEnvName)
	ev.rateLimit = envSet.Int(rateLimitEnvName)
	ev.cryptoKey = envSet.String(cryptoKeyEnvName)
	ev.configFile = envSet.String(configFileEnvName)
	return ev
}

func loadSettings(
	configFileFlag, configFileEnv *string,
	flagSet *flag.FlagSet, envSet *env.EnvSet,
) settings {
	settingsPath := getSettingsFilePath(
		configFileFlag, configFileEnv, flagSet, envSet,
	)

	if settingsPath != "" {
		settings, err := newSettings(settingsPath)
		if err != nil {
			failprint.Println(err)
			return settings
		}
		srcSettings = settingsPath
		return settings
	}
	return settings{}
}

func getSettingsFilePath(
	configFileFlag, configFileEnv *string,
	flagSet *flag.FlagSet, envSet *env.EnvSet,
) string {
	if envSet.IsSet(configFileEnvName) {
		return *configFileEnv
	}
	if flagSet.IsSet(configFileFlagName) {
		return *configFileFlag
	}
	return ""
}
