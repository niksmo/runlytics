package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

// Config describes server configurations parameters.
type Config struct {
	addr        *net.TCPAddr
	fileStorage *fileStorage
	database    db
	key         string
	logLvl      string
	isDatabase  bool
}

// Loag initializes flags and enviroments parameters then returns Config pointer.
func Load() *Config {
	rawLogLvlFlag := flag.String("l", logLvlDefault, logLvlUsage)
	rawAddrFlag := flag.String("a", addrDefault, addrUsage)
	rawFilePathFlag := flag.String("f", filePathDefault, filePathUsage)

	rawSaveIntervalFlag := flag.Int(
		"i",
		saveIntervalDefault,
		saveIntervalUsage,
	)

	rawRestoreFlag := flag.Bool("r", restoreDefault, restoreUsage)
	rawDatabaseDSNFlag := flag.String("d", databaseDSNDefault, databaseDSNUsage)
	rawKeyFlag := flag.String("k", keyDefault, keyUsage)
	flag.Parse()

	database := makeDatabaseConfig(*rawDatabaseDSNFlag)

	isDatabase := database.dsn != ""

	config := Config{
		logLvl: getLogLvlFlag(*rawLogLvlFlag),
		addr:   getAddrFlag(*rawAddrFlag),
		fileStorage: makeFileStorageConfig(
			!isDatabase,
			*rawFilePathFlag,
			*rawSaveIntervalFlag,
			*rawRestoreFlag,
		),
		database:   database,
		isDatabase: isDatabase,
		key:        getKeyFlag(*rawKeyFlag),
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
	if c.fileStorage.file != nil {
		file := *c.fileStorage.file
		return &file
	}
	return nil
}

// FileName returns filename of memory storage underlying file.
func (c *Config) FileName() string {
	if c.fileStorage.file != nil {
		return c.fileStorage.file.Name()
	}
	return ""
}

// SaveInterval returns memory storage save duration.
func (c *Config) SaveInterval() time.Duration {
	return c.fileStorage.saveInterval
}

// Restore returns memery storage restore flag.
func (c *Config) Restore() bool {
	return c.fileStorage.restore
}

// IsDatabase returns database usage flag.
func (c *Config) IsDatabase() bool {
	return c.isDatabase
}

// DatabaseDNS returns database data source name
func (c *Config) DatabaseDSN() string {
	return c.database.dsn
}

// Key returns hash checking key.
func (c *Config) Key() string {
	return c.key
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
