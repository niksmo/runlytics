package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	filePathEnv   = "FILE_STORAGE_PATH"
	filePathUsage = "Absolute file storage path, e.g. '/folder/file.ext'"

	saveIntervalDefault = 300
	saveIntervalUsage   = "File storage save interval, '0' is sync"
	saveIntervalEnv     = "STORE_INTERVAL"

	restoreDefault = true
	restoreUsage   = "Restore data from storage before start the server"
	restoreEnv     = "RESTORE"
)

var filePathDefault = getDefaultFilePath()

type fileStorage struct {
	file         *os.File
	saveInterval time.Duration
	restore      bool
}

func makeFileStorageConfig(
	isUsed bool,
	storagePath string,
	saveInterval int,
	restore bool,
) fileStorage {
	if !isUsed {
		return fileStorage{}
	}

	return fileStorage{
		file:         getFilePathFlag(storagePath),
		saveInterval: getSaveIntervalFlag(saveInterval),
		restore:      getRestoreFlag(restore),
	}
}

func getDefaultFilePath() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(execPath), "storage.json")
}

func getFilePathFlag(path string) *os.File {
	if envValue := os.Getenv(filePathEnv); envValue != "" {
		path = envValue
	}

	openFile := func() *os.File {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic("Open file storage error: " + err.Error())
		}
		return file
	}

	fileInfo, err := os.Stat(path)

	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0744); err != nil {
			panic("Create file storage path directories error: " + err.Error())
		}
		return openFile()
	}

	if err != nil {
		panic("File storage path error: " + err.Error())
	}

	if !fileInfo.Mode().IsRegular() {
		panic("File storage mode error: " + fileInfo.Mode().String())
	}

	return openFile()
}

func getSaveIntervalFlag(interval int) time.Duration {
	printError := func(isEnv bool, text string) {
		printParamError(isEnv, saveIntervalEnv, "-i", text)
	}
	printDefault := func() {
		printUsedDefault(
			"file storage save interval",
			fmt.Sprintf("%v", saveIntervalDefault),
		)
	}

	isEnv := false
	if envValue := os.Getenv(saveIntervalEnv); envValue != "" {
		isEnv = true
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			printError(isEnv, "error: value should be integer")
			interval = saveIntervalDefault
			printDefault()
		} else {
			interval = intValue
		}
	}

	if interval < 0 {
		text := "error: interval should be more or equal '0'"
		printError(isEnv, text)
		interval = saveIntervalDefault
		printDefault()
	}

	return time.Duration(interval) * time.Second
}

func getRestoreFlag(restore bool) bool {
	if envValue := os.Getenv(restoreEnv); envValue != "" {
		envValue = strings.ToLower(strings.TrimSpace(envValue))
		switch envValue {
		case "true", "1":
			restore = true
		case "false", "0":
			restore = false
		default:
			printParamError(
				true,
				restoreEnv,
				"-r", "parse error, expected values: true(1) or false(0)",
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
