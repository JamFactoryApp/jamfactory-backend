package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	log "github.com/sirupsen/logrus"
	"math/big"
	"os"
	"time"
)

func GenCerts(dnsNames []string, certFile, keyFile string) {
	log.Warn("Generating self-signed certificates")

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		log.Fatal("Failed to generate private key: ", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := crand.Int(crand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatal("Failed to generate serial number: ", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"JamFactoryApp"},
		},
		DNSNames:              dnsNames,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(crand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal("Failed to create certificate: ", err)
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		log.Fatal("Failed to encode certificate to PEM")
	}
	if err := os.WriteFile(certFile, pemCert, 0644); err != nil {
		log.Fatalf("Failed to save %s: %v\n", certFile, err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatal("Failed to marshal private key: ", err)
	}

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		log.Fatal("Failed to encode private key to PEM")
	}
	if err := os.WriteFile(keyFile, pemKey, 0600); err != nil {
		log.Fatalf("Failed to save %s: %v\n", keyFile, err)
	}
}
