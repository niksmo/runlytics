package config

import "os"

const (
	databaseDSNDefault = ""
	databaseDSNEnv     = "DATABASE_DSN"
	databaseDSNUsage   = "Usage 'postgres://user_name:user_pwd@localhost:5432/db_name?sslmode=disable'"
)

type db struct {
	dsn string
}

func makeDatabaseConfig(dsn string) db {
	if envValue := os.Getenv(databaseDSNEnv); envValue != "" {
		dsn = envValue
	}

	return db{dsn: dsn}

}
