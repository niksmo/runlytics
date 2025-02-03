package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type Config struct {
	logLvl      string
	addr        *net.TCPAddr
	fileStorage fileStorage
	database    database
	useDatabase bool
}

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

	flag.Parse()

	useDatabase := *rawDatabaseDSNFlag != ""

	config := Config{
		logLvl: getLogLvlFlag(*rawLogLvlFlag),
		addr:   getAddrFlag(*rawAddrFlag),
		fileStorage: makeFileStorageConfig(
			!useDatabase,
			*rawFilePathFlag,
			*rawSaveIntervalFlag,
			*rawRestoreFlag,
		),
		database:    makeDatabaseConfig(*rawDatabaseDSNFlag),
		useDatabase: useDatabase,
	}

	return &config
}

func (c *Config) LogLvl() string {
	return c.logLvl
}

func (c *Config) Addr() string {
	return c.addr.String()
}

func (c *Config) File() *os.File {
	file := *c.fileStorage.file
	return &file
}

func (c *Config) SaveInterval() time.Duration {
	return c.fileStorage.saveInterval
}

func (c *Config) Restore() bool {
	return c.fileStorage.restore
}

func (c *Config) UseDatabase() bool {
	return c.useDatabase
}

func (c *Config) DatabaseDSN() string {
	return c.database.dsn
}

func printUsedDefault(configField, value string) {
	fmt.Println("Used default", configField+":", value)
}

func printParamError(isEnv bool, envP, cliP, errText string) {
	var prefix string
	var p string
	if isEnv {
		prefix = "Env param"
		p = envP
	} else {
		prefix = "CLI param"
		p = cliP
	}
	fmt.Println(prefix, p, errText)
}
