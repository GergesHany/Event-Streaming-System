package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type TLSConfig struct {
	CertFile      string // Client or Server Certificate
	KeyFile       string // Client or Server Key
	CAFile        string // Certificate Authority
	ServerAddress string // Server address
	Server        bool // Is this config for a server?
}

func SetupTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}

	// Load client certificate
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(
			cfg.CertFile,
			cfg.KeyFile,
		)

		if err != nil {
			return nil, err
		}
	}

	// Load CA (Certificate Authority) certificate
	if cfg.CAFile != "" {
		b, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}

		// Create a certificate pool from the CA
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM([]byte(b))

		if !ok {
			return nil, fmt.Errorf("failed to parse root certificate: %q", cfg.CAFile)
		}

		// Set up TLS configuration
		if cfg.Server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.RootCAs = ca
		}

		tlsConfig.ServerName = cfg.ServerAddress
	}

	return tlsConfig, nil
}
