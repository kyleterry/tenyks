package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/kyleterry/tenyks/internal/config"
)

// CAAndCertPair holds a CA certificate and a certificate signed by the CA.
type CAAndCertPair struct {
	// CA is the certificate authority used to verify peer certificates.
	CA *x509.Certificate
	// Cert is the certificate signed by the CA. The private key is attached.
	Cert *tls.Certificate
}

// Load reads the CA certificate and the cert/key pair described by cfg.
func Load(cfg *config.TLS) (*CAAndCertPair, error) {
	caPEMBytes, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca certificate file: %w", err)
	}

	caBlock, _ := pem.Decode(caPEMBytes)
	if caBlock == nil {
		return nil, fmt.Errorf("failed to decode ca certificate: invalid pem block")
	}

	ca, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca certificate: %w", err)
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	return &CAAndCertPair{
		CA:   ca,
		Cert: &cert,
	}, nil
}

// NewServerConfig returns a *tls.Config suitable for an mTLS server.
// It presents certs.Cert to connecting clients and requires every client to
// present a certificate signed by certs.CA.
func NewServerConfig(certs *CAAndCertPair) *tls.Config {
	pool := x509.NewCertPool()
	pool.AddCert(certs.CA)

	return &tls.Config{
		ClientCAs:    pool,
		Certificates: []tls.Certificate{*certs.Cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}
}

// NewClientConfig returns a *tls.Config suitable for an mTLS client.
// It presents certs.Cert to the server and verifies the server certificate
// against certs.CA.
func NewClientConfig(certs *CAAndCertPair) *tls.Config {
	pool := x509.NewCertPool()
	pool.AddCert(certs.CA)

	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{*certs.Cert},
		MinVersion:   tls.VersionTLS13,
	}
}
