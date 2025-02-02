package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type Config struct {
	logLvl       string
	addr         *net.TCPAddr
	filePath     *os.File
	saveInterval time.Duration
	restore      bool
	dbDSN        string
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

	config := Config{
		logLvl:       getLogLvlFlag(*rawLogLvlFlag),
		addr:         getAddrFlag(*rawAddrFlag),
		filePath:     getFilePathFlag(*rawFilePathFlag),
		saveInterval: getSaveIntervalFlag(*rawSaveIntervalFlag),
		restore:      getRestoreFlag(*rawRestoreFlag),
		dbDSN:        getDatabaseDSNFlag(*rawDatabaseDSNFlag),
	}

	return &config
}

func (c *Config) LogLvl() string {
	return c.logLvl
}

func (c *Config) Addr() string {
	return c.addr.String()
}

func (c *Config) StoragePath() *os.File {
	file := *c.filePath
	return &file
}

func (c *Config) SaveInterval() time.Duration {
	return c.saveInterval
}

func (c *Config) Restore() bool {
	return c.restore
}

func (c *Config) DatabaseDSN() string {
	return c.dbDSN
}

func printEnvParamError(p, errText string) {
	fmt.Println("Env param", p, errText)
}

func printCliParamError(p, errText string) {
	fmt.Println("Cli param", p, errText)
}

func printUsedDefault(configField, value string) {
	fmt.Println("Used default", configField+":", value)
}
