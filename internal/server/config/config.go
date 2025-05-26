package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	trustedNetFlagName     = "t"
	trustedNetEnvName      = "TRUSTED_SUBNET"
	trustedNetSettingsName = "trusted_subnet"
	trustedNetDefault      = ""
	trustedNetUsage        = "Trusted subnet CIDR, e.g. '192.168.1.1/24'"

	configFileFlagName = "config"
	configFileEnvName  = "CONFIG"
	configFileDefault  = ""
	configFileUsage    = "Path to json config file, e.g. '/folder/to/config.json' (optional)"
)

var storeDefaultPath = getStoreDefaultPath()

type values struct {
	addr          *string
	log           *string
	dsn           *string
	store         *string
	storeInterval *int
	storeRestore  *bool
	hashKey       *string
	cryptoKey     *string
	trustedNet    *string
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
	TrustedNet    *string `json:"trusted_subnet"`
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

// ServerConfig describes server configurations parameters.
type ServerConfig struct {
	Addr        AddrConfig
	FileStorage FileStorageConfig
	Log         LogConfig
	DB          DBConfig
	HashKey     HashKeyConfig
	Crypto      CryptoConfig
	TrustedNet  TrustedNetConfig
}

// Loag initializes flags and enviroments parameters then returns Config pointer.
func Load() *ServerConfig {
	var (
		flagSet           = flag.New()
		envSet            = env.New()
		fileStorageConfig FileStorageConfig
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

	addrConfig := NewAddrConfig(params)
	logConfig := NewLogConfig(params)
	dbConfig := NewDBConfig(params)

	if !dbConfig.IsSet() {
		fileStorageConfig = NewFileStorageConfig(params)
	}

	hashKeyConfig := NewHashKeyConfig(params)
	cryptoConfig := NewCryptoConfig(params)
	trustedNetConfig := NewTrustedNetConfig(params)

	return &ServerConfig{
		Addr:        addrConfig,
		Log:         logConfig,
		FileStorage: fileStorageConfig,
		DB:          dbConfig,
		HashKey:     hashKeyConfig,
		Crypto:      cryptoConfig,
		TrustedNet:  trustedNetConfig,
	}
}

// IsDatabase returns database usage flag.
func (c *ServerConfig) IsDatabase() bool {
	return c.DB.IsSet()
}

func setupFlagValues(flagSet *flag.FlagSet) values {
	var fv values

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
	fv.trustedNet = flagSet.String(
		trustedNetFlagName, trustedNetDefault, trustedNetUsage,
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
	ev.store = envSet.String(storeEnvName)
	ev.storeInterval = envSet.Int(storeIntervalEnvName)
	ev.storeRestore = envSet.Bool(storeRestoreEnvName)
	ev.dsn = envSet.String(dsnEnvName)
	ev.hashKey = envSet.String(hashKeyEnvName)
	ev.cryptoKey = envSet.String(cryptoKeyEnvName)
	ev.trustedNet = envSet.String(trustedNetEnvName)
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
