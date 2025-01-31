package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type addr struct {
	host string
	port int
}

func (a *addr) String() string {
	return a.host + ":" + strconv.Itoa(a.port)
}

func (a *addr) Set(v string) error {
	var errAddr = errors.New("incorrect address format, usage: example.com:8080")
	slice := strings.SplitN(v, ":", 2)
	if len(slice) != 2 {
		return errAddr
	}

	port, err := strconv.Atoi(slice[1])

	if err != nil {
		return errAddr
	}

	a.host = slice[0]
	a.port = port
	return nil
}

var (
	flagAddr        *addr = &addr{host: "localhost", port: 8080}
	flagLog         string
	flagInterval    time.Duration
	flagStoragePath *os.File
	flagRestore     bool
	flagDSN         string
)

func parseFlags() {
	// init flags
	defaultStoragePath, err := os.UserHomeDir()
	if err != nil {
		log.Panic("Users home path environment not set")
	}
	defaultStoragePath = filepath.Join(defaultStoragePath, "runlytics", "storage.json")

	flag.Var(flagAddr, "a", "Listening server address, e.g. example.com:8080")
	flag.StringVar(&flagLog, "l", "info", "Logging level, e.g. debug")
	rawFlagInterval := flag.Int("i", 300, "Storage save interval, '0' is sync")
	rawFlagStoragePath := flag.String("f", defaultStoragePath, "Absolute path for storage, e.g. /folder/file.ext")
	flag.BoolVar(&flagRestore, "r", true, "Restore data from storage before runnig server")
	flag.StringVar(&flagDSN, "d", "", "Usage \"postgres://user_name:user_pwd@localhost:5432/db_name?sslmode=disable\"")
	flag.Parse()

	// get env
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		if err := flagAddr.Set(envAddr); err != nil {
			log.Print(fmt.Errorf("parse env ADDRESS error: %w", err))
		}
	}

	if envLog := os.Getenv("LOG_LVL"); envLog != "" {
		flagLog = envLog
	}

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		interval, err := strconv.Atoi(envInterval)
		if err != nil {
			log.Print(fmt.Errorf("parse env STORE_INTERVAL error: %w", err))
		} else {
			rawFlagInterval = &interval
		}
	}

	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		rawFlagStoragePath = &envStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		envRestore = strings.ToLower(strings.TrimSpace(envRestore))
		switch envRestore {
		case "true", "1":
			flagRestore = true
		case "false", "0":
			flagRestore = false
		default:
			log.Print("parse env RESTORE error, expected values: true(1), false(0)")
		}
	}

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		flagDSN = envDSN
	}

	// convert and validate raw flags
	flagInterval = time.Second * time.Duration(*rawFlagInterval)

	if err := setFlagStoragePath(*rawFlagStoragePath); err != nil {
		log.Panic(fmt.Errorf("storage path: %w", err))
	}

}

func setFlagStoragePath(path string) error {
	fileInfo, err := os.Stat(path)

	openFile := func() error {
		flagStoragePath, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("can't open storage file path: %w", err)
		}
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0744); err != nil {
			return fmt.Errorf("can't create path directories")
		}
		if err := openFile(); err != nil {
			return err
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("unexpected storage file path: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("unexpected storage file mode: %s", fileInfo.Mode())
	}

	if err := openFile(); err != nil {
		return err
	}
	return nil
}
