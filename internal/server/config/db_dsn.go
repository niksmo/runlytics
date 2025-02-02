package config

import "os"

const (
	dbDSNDefault = ""
	dbDSNEnv     = "DATABASE_DSN"
	dbDSNUsage   = "Usage 'postgres://user_name:user_pwd@localhost:5432/db_name?sslmode=disable'"
)

func getDbDSNFlag(dsn string) string {
	if envValue := os.Getenv(dbDSNEnv); envValue != "" {
		dsn = envValue
	}
	return dsn
}
