package config

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	cryptoKeyDefault = ""
	cryptoKeyEnv     = "CRYPTO_KEY"
	cryptoKeyUsage   = "Private key absolute path, e.g. '/folder/key.pem' (required)"
)

var (
	ErrCryptoKeyEmptyFlag = errors.New("required -crypto-key flag not set")
	ErrCryptoKeyOpenFile  = errors.New("failed to open crypto key file")
	ErrCryptoKeyReadFile  = errors.New("failed to read crypto key file")
)

type cryptoKey struct {
	path    string
	pemData []byte
}

func getCryptoKeyFlag(path string) cryptoKey {
	if envValue := os.Getenv(cryptoKeyEnv); envValue != "" {
		path = envValue
	}

	if path == "" {
		fmt.Printf("%s\ntype --help for usage\n", ErrCryptoKeyEmptyFlag)
		os.Exit(1)
	}

	pemData, err := getPEMData(path)
	if err != nil {
		panic(err)
	}
	return cryptoKey{path, pemData}
}

func getPEMData(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Join(ErrCryptoKeyOpenFile, err)
	}
	pemData, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Join(ErrCryptoKeyReadFile, err)
	}
	return pemData, nil
}
