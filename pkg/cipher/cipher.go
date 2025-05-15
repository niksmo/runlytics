// Package cipher provides Encrypter and Decrypter over X.509 standart.
package cipher

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
)

var (
	ErrFileLoad        = errors.New("failed to load file")
	ErrParseCert       = errors.New("failed to parse certificate")
	ErrParsePrivateKey = errors.New("failed to parse private key")
	ErrPublicKeyType   = errors.New("invalid RSA PublicKey")
	ErrEncryptMsg      = errors.New("failed to encrypt message")
	ErrDecryptMsg      = errors.New("failed to decrypt message")
)

// Encrypter provides simple message encrypting.
type Encrypter struct {
	publicKey *rsa.PublicKey
	hash      hash.Hash
}

// NewEncrypter returns Encrypter pointer.
func NewEncrypter(PEMData []byte) (*Encrypter, error) {
	block, _ := pem.Decode(PEMData)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Join(ErrParseCert, err)
	}
	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrPublicKeyType
	}
	ecrypter := &Encrypter{publicKey: rsaPublicKey, hash: sha256.New()}
	return ecrypter, nil
}

// EncryptMsg returns ciphered message data.
func (e *Encrypter) EncryptMsg(msg []byte) ([]byte, error) {
	data, err := rsa.EncryptOAEP(e.hash, rand.Reader, e.publicKey, msg, nil)
	if err != nil {
		return nil, errors.Join(ErrEncryptMsg, err)
	}
	return data, nil
}

// Decrypter provides simple message decrypting.
type Decrypter struct {
	privateKey *rsa.PrivateKey
	hash       hash.Hash
}

// NewDecrypter returns Decrypter pointer.
func NewDecrypter(PEMData []byte) (*Decrypter, error) {
	block, _ := pem.Decode(PEMData)

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Join(ErrParsePrivateKey, err)
	}
	decrypter := &Decrypter{privateKey: rsaPrivateKey, hash: sha256.New()}
	return decrypter, nil
}

// DecryptMsg returns unciphered message data.
func (d *Decrypter) DecryptMsg(msg []byte) ([]byte, error) {
	data, err := rsa.DecryptOAEP(d.hash, rand.Reader, d.privateKey, msg, nil)
	if err != nil {
		return nil, errors.Join(ErrDecryptMsg, err)
	}
	return data, nil
}
