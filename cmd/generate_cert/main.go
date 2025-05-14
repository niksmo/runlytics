package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	oneYear          = 365 * 24 * time.Hour
	privateKeyLength = 4096
)

func main() {
	certTemplate, err := makeTemplate()
	if err != nil {
		log.Fatalln(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, privateKeyLength)
	if err != nil {
		log.Fatal(err)
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		certTemplate,
		certTemplate,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		log.Fatalf("failed to create certificate: %v", err)
	}

	err = saveCert(derBytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrote cert.pem")

	err = savePrivateKey(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("wrote key.pem")
}

func generateSerial() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return serialNumber, nil
}

func makeTemplate() (*x509.Certificate, error) {
	serialNumber, err := generateSerial()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	subject := pkix.Name{
		Organization: []string{"Runlytics"},
		Country:      []string{"RU"},
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment

	notBefore := time.Now()
	notAfter := notBefore.Add(oneYear)

	ipAddresses := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}

	extKeyUsage := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
		x509.ExtKeyUsageServerAuth,
	}

	certTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IPAddresses:           ipAddresses,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}
	return &certTemplate, nil
}

func saveCert(derBytes []byte) error {
	certOut, err := os.Create("cert.pem")
	if err != nil {
		return fmt.Errorf("failed to open cert.pem for writing: %w", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write data to cert.pem: %w", err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("error closing cert.pem: %w", err)
	}

	return nil
}

func savePrivateKey(key *rsa.PrivateKey) error {
	keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open key.pem for writing: %w", err)
	}

	err = pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return fmt.Errorf("failed to write data to key.pem: %w", err)
	}
	return nil
}
