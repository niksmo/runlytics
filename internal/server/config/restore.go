package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	restoreDefault = true
	restoreUsage   = "Restore data from storage before start the server"
	restoreEnv     = "RESTORE"
)

func getRestoreFlag(restore bool) bool {
	if envValue := os.Getenv(restoreEnv); envValue != "" {
		envValue = strings.ToLower(strings.TrimSpace(envValue))
		switch envValue {
		case "true", "1":
			restore = true
		case "false", "0":
			restore = false
		default:
			printEnvParamError(
				restoreEnv,
				"parse error, expected values: true(1) or false(0)",
			)
			restore = restoreDefault
			printUsedDefault(
				"restore file storage",
				fmt.Sprintf("%v", restoreDefault),
			)
		}
	}

	return restore

}
