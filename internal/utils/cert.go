package utils

import (
	"crypto/tls"
	"crypto/x509"
	"openshield-manager/internal/config"
	"os"
)

func LoadClientTLSCredentials() (*tls.Config, error) {
	// Load TLS credentials from the provided files
	cert, err := tls.LoadX509KeyPair(config.CertsPath+"/manager.crt", config.CertsPath+"/manager.key")
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(config.CertsPath + "/ca.crt")
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}, nil

}

func LoadServerTLSCredentials() (*tls.Config, error) {
	// Load TLS credentials from the provided files
	cert, err := tls.LoadX509KeyPair(config.CertsPath+"/manager.crt", config.CertsPath+"/manager.key")
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(config.CertsPath + "/ca.crt")
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
