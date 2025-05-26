// Package cipher provides Encrypter and Decrypter over X.509 standart.
package cipher

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	ErrFileLoad        = errors.New("failed to load file")
	ErrParseCert       = errors.New("failed to parse certificate")
	ErrParsePrivateKey = errors.New("failed to parse private key")
	ErrPublicKeyType   = errors.New("invalid RSA PublicKey")
	ErrEncryptMsg      = errors.New("failed to encrypt message")
)

type (
	Encrypter interface {
		EncryptMsg([]byte) ([]byte, error)
	}

	Decrypter interface {
		DecryptMsg([]byte) ([]byte, error)
	}
)

// Encrypter provides simple message encrypting.
type encrypterX509 struct {
	publicKey *rsa.PublicKey
}

// NewEncrypter returns Encrypter pointer.
func NewEncrypterX509(PEMData []byte) (Encrypter, error) {
	block, _ := pem.Decode(PEMData)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Join(ErrParseCert, err)
	}
	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrPublicKeyType
	}
	encrypter := &encrypterX509{publicKey: rsaPublicKey}
	return encrypter, nil
}

// EncryptMsg returns ciphered message data.
func (e *encrypterX509) EncryptMsg(msg []byte) ([]byte, error) {
	data, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, msg, nil)
	if err != nil {
		return nil, errors.Join(ErrEncryptMsg, err)
	}
	return data, nil
}

// Decrypter provides simple message decrypting.
type decrypterX509 struct {
	privateKey *rsa.PrivateKey
}

// NewDecrypter returns Decrypter pointer.
// If PEMData is invalid error in occur.
func NewDecrypterX509(PEMData []byte) (*decrypterX509, error) {
	block, _ := pem.Decode(PEMData)

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Join(ErrParsePrivateKey, err)
	}
	decrypter := &decrypterX509{privateKey: rsaPrivateKey}
	return decrypter, nil
}

// DecryptMsg returns unciphered message data.
func (d *decrypterX509) DecryptMsg(msg []byte) ([]byte, error) {
	data, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privateKey, msg, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}
