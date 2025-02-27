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
	fileStorage *fileStorage
	database    database
	isDatabase  bool
	key         string
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

func (c *Config) LogLvl() string {
	return c.logLvl
}

func (c *Config) Addr() string {
	return c.addr.String()
}

func (c *Config) File() *os.File {
	if c.fileStorage.file != nil {
		file := *c.fileStorage.file
		return &file
	}
	return nil
}
func (c *Config) FileName() string {
	if c.fileStorage.file != nil {
		return c.fileStorage.file.Name()
	}
	return ""
}

func (c *Config) SaveInterval() time.Duration {
	return c.fileStorage.saveInterval
}

func (c *Config) Restore() bool {
	return c.fileStorage.restore
}

func (c *Config) IsDatabase() bool {
	return c.isDatabase
}

func (c *Config) DatabaseDSN() string {
	return c.database.dsn
}

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
