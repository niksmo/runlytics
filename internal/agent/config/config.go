package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/failprint"
	"github.com/niksmo/runlytics/pkg/flag"
	"go.uber.org/zap"
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
	addrDefault      = "localhost:8080"
	addrUsage        = "TCP address for metrics emitting, e.g. '192.168.1.101:8080'"

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

type values struct {
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

type ConfigParams struct {
	FlagValues, EnvValues values
	FlagSet               *flag.FlagSet
	EnvSet                *env.EnvSet
	Settings              settings
	ErrStream             chan<- error
}

type AgentConfig struct {
	Server  ServerConfig
	Log     LogConfig
	Metrics MetricsConfig
	HashKey HashKeyConfig
	Crypto  CryptoConfig
}

func Load() *AgentConfig {
	var (
		flagSet = flag.New()
		envSet  = env.New()
	)
	flagV := setupFlagValues(flagSet)
	envV := setupEnvValues(envSet)

	flagSet.Parse()
	if err := envSet.Parse(); err != nil {
		failprint.Println(err)
	}

	settings := loadSettings(flagV.configFile, envV.configFile, flagSet, envSet)

	errStream := make(chan error)
	defer close(errStream)
	go failprint.PrintFailWorker(errStream, failprint.ExitOnError)

	params := ConfigParams{
		FlagValues: flagV,
		EnvValues:  envV,
		FlagSet:    flagSet,
		EnvSet:     envSet,
		Settings:   settings,
		ErrStream:  errStream,
	}

	serverConfig := NewServerConfig(params)
	logConfig := NewLogConfig(params)
	metricsConfig := NewMetricsConfig(params)
	hashKeyConfig := NewHashKeyConfig(params)
	cryptoConfig := NewCryptoConfig(params)

	return &AgentConfig{
		Server:  serverConfig,
		Log:     logConfig,
		Metrics: metricsConfig,
		HashKey: hashKeyConfig,
		Crypto:  cryptoConfig,
	}

}

func (c *AgentConfig) GetOutboundIP() string {
	conn, err := net.Dial("udp", c.Server.Addr.String())
	if err != nil {
		failprint.Println(err)
		os.Exit(2)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (c *AgentConfig) PrintConfig(logger *zap.Logger) {
	logger.Info(
		"Start agent with flags",
		zap.String("-"+addrFlagName, c.Server.URL()),
		zap.String("-"+logFlagName, c.Log.Level),
		zap.String("-"+pollFlagName, c.Metrics.Poll.String()),
		zap.String("-"+reportFlagName, c.Metrics.Report.String()),
		zap.String("-"+hashKeyFlagName, c.HashKey.Key),
		zap.Int("-"+rateLimitFlagName, c.Metrics.RateLimit),
		zap.String("-"+cryptoKeyFlagName, c.Crypto.Path),
		zap.String("outboundIP", c.GetOutboundIP()),
	)
}

func setupFlagValues(flagSet *flag.FlagSet) values {
	var fv values

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

func setupEnvValues(envSet *env.EnvSet) values {
	var ev values
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
