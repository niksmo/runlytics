package config

import "os"

const (
	keyDefault = ""
	keyUsage   = "Sercret key for hashing." +
		" Set HashSHA256 value in HTTP header requests"
	keyEnv = "KEY"
)

func getKeyFlag(key string) string {
	if envValue := os.Getenv(keyEnv); envValue != "" {
		return envValue
	}
	return key
}
