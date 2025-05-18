package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
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

type flagValues struct {
	addr          *string
	log           *string
	dsn           *string
	store         *string
	storeInterval *int
	storeRestore  *bool
	hashKey       *string
	cryptoKey     *string
	configFile    *string
}

type envValues struct {
	addr          *string
	log           *string
	dsn           *string
	store         *string
	storeInterval *int
	storeRestore  *bool
	hashKey       *string
	cryptoKey     *string
	configFile    *string
}

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
	cryptoKeyFile    *os.File
	cryptoKeyPEMData []byte
}

// Loag initializes flags and enviroments parameters then returns Config pointer.
func Load() *Config {
	var (
		flagSet         = flag.New()
		envSet          = env.New()
		storeFileConfig *os.File
		cryptoKeyData   []byte
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

	dsnConfig := getDSNConfig(flagV.dsn, envV.dsn, flagSet, envSet, settings)

	if dsnConfig == "" {
		storeFileConfig = getStoreFileConfig(
			flagV.store, envV.store, flagSet, envSet, settings, errStream,
		)
	}

	storeIntervalConfig := getStoreIntervalConfig(
		flagV.storeInterval,
		envV.storeInterval,
		flagSet,
		envSet,
		settings,
		errStream,
	)

	storeRestoreConfig := getStoreRestoreConfig(
		flagV.storeRestore, envV.storeRestore, flagSet, envSet, settings,
	)

	hashKeyConfig := getHashKeyConfig(
		flagV.hashKey, envV.hashKey, flagSet, envSet, settings,
	)

	cryptoKeyFile := getCryptoKeyFile(
		flagV.cryptoKey, envV.cryptoKey, flagSet, envSet, settings, errStream,
	)

	if cryptoKeyFile != nil {
		cryptoKeyData = getCryptoKeyData(cryptoKeyFile, errStream)
	}

	close(errStream)

	config := Config{
		addr:             addrConfig,
		logLvl:           logConfig,
		storeFile:        storeFileConfig,
		storeInterval:    storeIntervalConfig,
		storeRestore:     storeRestoreConfig,
		dsn:              dsnConfig,
		hashKey:          hashKeyConfig,
		cryptoKeyFile:    cryptoKeyFile,
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
	if c.cryptoKeyFile != nil {
		return c.cryptoKeyFile.Name()
	}
	return ""
}

// CryptoKeyData returns private key PEM file data.
func (c *Config) CryptoKeyData() []byte {
	return c.cryptoKeyPEMData
}

func setupFlagValues(flagSet *flag.FlagSet) flagValues {
	var fv flagValues

	fv.addr = flagSet.String(addrFlagName, addrDefault, addrUsage)
	fv.log = flagSet.String(logFlagName, logDefault, logUsage)

	fv.store = flagSet.String(
		storeFlagName, "cwd/store.json", storeUsage,
	)

	fv.storeInterval = flagSet.Int(
		storeIntervalFlagName,
		storeIntervalDefault,
		storeIntervalUsage,
	)

	fv.storeRestore = flagSet.Bool(
		storeRestoreFlagName, storeRestoreDefault, storeRestoreUsage,
	)

	fv.dsn = flagSet.String(dsnFlagName, dsnDefault, dsnUsage)
	fv.hashKey = flagSet.String(hashKeyFlagName, hashKeyDefault, hashKeyUsage)

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
	ev.store = envSet.String(storeEnvName)
	ev.storeInterval = envSet.Int(storeIntervalEnvName)
	ev.storeRestore = envSet.Bool(storeRestoreEnvName)
	ev.dsn = envSet.String(dsnEnvName)
	ev.hashKey = envSet.String(hashKeyEnvName)
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

func getStoreDefaultPath() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(execPath), "store.json")
}
