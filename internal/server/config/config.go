package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/niksmo/runlytics/pkg/flag"
)

const (
	srcEnv      = "env"
	srcFlag     = "arg"
	srcSettings = "config"
)

const (
	addrFlagName     = "a"
	addrEnvName      = "ADDRESS"
	addrSettingsName = "address"
	addrDefault      = "localhost:8080"
	addrUsage        = "Listening server address, e.g. '127.0.0.1:8080'"

	logFlagName     = "l"
	logEnvName      = "LOG_LVL"
	logSettingsName = "log"
	logDefault      = "info"
	logUsage        = "Logging level, e.g. 'debug'"

	storeFlagName             = "f"
	storeEnvName              = "FILE_STORAGE_PATH"
	storeIntervalSettingsName = "store_interval"
	storeSettingsName         = "store_file"
	storeUsage                = "Path to storage file, e.g. '/folder/to/file.ext'"

	storeIntervalFlagName = "i"
	storeIntervalEnvName  = "STORE_INTERVAL"
	storeIntervalDefault  = 300
	storeIntervalUsage    = "File storage save interval, '0' is sync"

	storeRestoreFlagName     = "r"
	storeRestoreEnvName      = "RESTORE"
	storeRestoreSettingsName = "restore"
	storeRestoreDefault      = true
	storeRestoreUsage        = "Restore data from storage before start the server"

	dsnFlagName = "d"
	dsnEnvName  = "DATABASE_DSN"
	dsnDefault  = ""
	dsnUsage    = "Data source 'postgres://user_name:user_pwd@localhost:5432/db_name?sslmode=disable' (optional)"

	hashKeyFlagName = "k"
	hashKeyEnvName  = "KEY"
	hashKeyDefault  = ""
	hashKeyUsage    = "Key for verify and pass hash in HTTP header (optional)"

	cryptoKeyFlagName     = "crypto-key"
	cryptoKeyEnvName      = "CRYPTO_KEY"
	cryptoKeySettingsName = "crypto_key"
	cryptoKeyDefault      = ""
	cryptoKeyUsage        = "Private key absolute path, e.g. '/folder/key.pem' (required)"

	configFileFlagName = "config"
	configFileEnvName  = "CONFIG"
	configFileDefault  = ""
	configFileUsage    = "Path to json config file, e.g. '/folder/to/config.json' (optional)"
)

var storeDefaultPath = getStoreDefaultPath()

type settings struct {
	Address       *string `json:"address"`
	Log           *string `json:"log"`
	StoreFile     *string `json:"store_file"`
	StoreInterval *int    `json:"store_interval"`
	Restore       *bool   `json:"restore"`
	DSN           *string `json:"database_dsn"`
	HashKey       *string `json:"hash_key"`
	CryptoKey     *string `json:"crypto_key"`
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

// Config describes server configurations parameters.
type Config struct {
	addr             *net.TCPAddr
	logLvl           string
	storeFile        *os.File
	storeInterval    time.Duration
	storeRestore     bool
	dsn              string
	hashKey          string
	cryptoKeyPath    string
	cryptoKeyPEMData []byte
}

// Loag initializes flags and enviroments parameters then returns Config pointer.
func Load() *Config {
	var (
		flagSet = flag.New()
		envSet  = env.New()
	)

	addrFlag := flagSet.String(addrFlagName, addrDefault, addrUsage)
	addrEnv := envSet.String(addrEnvName)

	logFlag := flagSet.String(logFlagName, logDefault, logUsage)
	logEnv := envSet.String(logEnvName)

	storeFlag := flagSet.String(
		storeFlagName, storeDefaultPath, storeUsage,
	)
	storeEnv := envSet.String(storeEnvName)

	storeIntervalFlag := flagSet.Int(
		storeIntervalFlagName,
		storeIntervalDefault,
		storeIntervalUsage,
	)
	storeIntervalEnv := envSet.Int(storeIntervalEnvName)

	storeRestoreFlag := flagSet.Bool(
		storeRestoreFlagName, storeRestoreDefault, storeRestoreUsage,
	)
	storeRestoreEnv := envSet.Bool(storeRestoreEnvName)

	dsnFlag := flagSet.String(dsnFlagName, dsnDefault, dsnUsage)
	dsnEnv := envSet.String(dsnEnvName)

	hashKeyFlag := flagSet.String(hashKeyFlagName, hashKeyDefault, hashKeyUsage)
	hashKeyEnv := envSet.String(hashKeyEnvName)

	cryptoKeyFlag := flagSet.String(cryptoKeyFlagName, cryptoKeyDefault, cryptoKeyUsage)
	cryptoKeyEnv := envSet.String(cryptoKeyEnvName)

	configFileFlag := flagSet.String(configFileFlagName, configFileDefault, configFileUsage)
	configFileEnv := envSet.String(configFileEnvName)

	flagSet.Parse()
	if err := envSet.Parse(); err != nil {
		printFail(err)
	}

	settings := loadSettings(configFileFlag, configFileEnv, flagSet, envSet)

	errStream := make(chan error)
	go errorsWorker(errStream)

	addrConfig := getAddrConfig(
		addrFlag, addrEnv, flagSet, envSet, settings, errStream,
	)

	logConfig := getLogConfig(
		logFlag, logEnv, flagSet, envSet, settings, errStream,
	)

	dsnConfig := getDSNConfig(dsnFlag, dsnEnv, flagSet, envSet, settings)

	var storeFileConfig *os.File

	if dsnConfig == "" {
		storeFileConfig = getStoreFileConfig(
			storeFlag, storeEnv, flagSet, envSet, settings, errStream,
		)
	}

	storeIntervalConfig := getStoreIntervalConfig(
		storeIntervalFlag,
		storeIntervalEnv,
		flagSet,
		envSet,
		settings,
		errStream,
	)

	storeRestoreConfig := getStoreRestoreConfig(
		storeRestoreFlag, storeRestoreEnv, flagSet, envSet, settings,
	)

	hashKeyConfig := getHashKeyConfig(
		hashKeyFlag, hashKeyEnv, flagSet, envSet, settings,
	)

	cryptoKeyFile := getCryptoKeyFile(
		cryptoKeyFlag, cryptoKeyEnv, flagSet, envSet, settings, errStream,
	)

	cryptoKeyData := getCryptoKeyData(cryptoKeyFile, errStream)

	close(errStream)

	config := Config{
		addr:             addrConfig,
		logLvl:           logConfig,
		storeFile:        storeFileConfig,
		storeInterval:    storeIntervalConfig,
		storeRestore:     storeRestoreConfig,
		dsn:              dsnConfig,
		hashKey:          hashKeyConfig,
		cryptoKeyPath:    cryptoKeyFile.Name(),
		cryptoKeyPEMData: cryptoKeyData,
	}

	return &config
}

// LogLvl returns logging level.
func (c *Config) LogLvl() string {
	return c.logLvl
}

// Addr returns server listening address.
func (c *Config) Addr() string {
	return c.addr.String()
}

// File returns memory storage underlying file.
func (c *Config) File() *os.File {
	if c.storeFile != nil {
		file := *c.storeFile
		return &file
	}
	return nil
}

// FileName returns filename of memory storage underlying file.
func (c *Config) FileName() string {
	if c.storeFile != nil {
		return c.storeFile.Name()
	}
	return ""
}

// SaveInterval returns memory storage save duration.
func (c *Config) SaveInterval() time.Duration {
	return c.storeInterval
}

// Restore returns memery storage restore flag.
func (c *Config) Restore() bool {
	return c.storeRestore
}

// IsDatabase returns database usage flag.
func (c *Config) IsDatabase() bool {
	return c.dsn != ""
}

// DatabaseDNS returns database data source name
func (c *Config) DatabaseDSN() string {
	return c.dsn
}

// Key returns hash checking key.
func (c *Config) Key() string {
	return c.hashKey
}

// CryptoKeyPath returns private key path.
func (c *Config) CryptoKeyPath() string {
	return c.cryptoKeyPath
}

// CryptoKeyData returns private key PEM file data.
func (c *Config) CryptoKeyData() []byte {
	return c.cryptoKeyPEMData
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
			printFail(err)
		}
		return settings
	}
	return settings{}
}

func getSettingsFilePath(
	configFileFlag, configFileEnv *string,
	flagSet *flag.FlagSet, envSet *env.EnvSet,
) string {
	if flagSet.IsSet(configFileEnvName) {
		return *configFileEnv
	}
	if envSet.IsSet(configFileFlagName) {
		return *configFileFlag
	}
	return ""
}

func getStoreDefaultPath() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(execPath), "storage.json")
}

func printFail(err error) {
	fmt.Fprint(os.Stderr, err)
}

func errorsWorker(errStream <-chan error) {
	var shouldExit bool
	for err := range errStream {
		shouldExit = true
		printFail(err)
	}
	if shouldExit {
		os.Exit(2)
	}
}
