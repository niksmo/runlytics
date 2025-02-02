package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	filePathEnv   = "FILE_STORAGE_PATH"
	filePathUsage = "Absolute file storage path, e.g. '/folder/file.ext'"
)

var filePathDefault = getDefaultFilePath()

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

func getDefaultFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Users 'HOME' path environment not set")
	}
	return filepath.Join(homeDir, "runlytics", "storage.json")
}
