package certutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// GenerateClientCert creates a new ECDSA P-256 client certificate signed by
// the CA at caCertPath/caKeyPath, embedding perms as a custom extension.
// Returns PEM-encoded certificate and private key bytes.
func GenerateClientCert(name, caCertPath, caKeyPath string, perms Permissions, days int) (certPEM, keyPEM []byte, err error) {
	caCert, caKey, err := loadCA(caCertPath, caKeyPath)
	if err != nil {
		return nil, nil, err
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	permsExts, err := EncodeExtensions(perms)
	if err != nil {
		return nil, nil, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("generate serial: %w", err)
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber:    serial,
		Subject:         pkix.Name{CommonName: name},
		NotBefore:       now,
		NotAfter:        now.Add(time.Duration(days) * 24 * time.Hour),
		KeyUsage:        x509.KeyUsageDigitalSignature,
		ExtKeyUsage:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		ExtraExtensions: permsExts,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("sign certificate: %w", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal key: %w", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

func loadCA(certPath, keyPath string) (*x509.Certificate, crypto.Signer, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read ca cert: %w", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read ca key: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("no PEM block in ca cert")
	}
	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse ca cert: %w", err)
	}

	caKey, err := parsePrivateKey(keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse ca key: %w", err)
	}

	return caCert, caKey, nil
}

func parsePrivateKey(pemBytes []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in key file")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key does not implement crypto.Signer")
		}
		return signer, nil
	default:
		return nil, fmt.Errorf("unsupported key PEM type: %s", block.Type)
	}
}
