package config

import "os"

const (
	keyDefault = ""
	keyUsage   = "Sercret key for verify and pass hash in HTTP header."
	keyEnv     = "KEY"
)

func getKeyFlag(key string) string {
	if envValue := os.Getenv(keyEnv); envValue != "" {
		return envValue
	}
	return key
}
