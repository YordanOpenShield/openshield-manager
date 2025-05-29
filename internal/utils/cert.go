package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"openshield-manager/internal/config"
	"os"
	"time"
)

// Generate a CA key and certificate
func GenerateCA(caCN string) error {
	// Check if CA key and cert already exist
	keyPath := config.CertsPath + "/ca.key"
	certPath := config.CertsPath + "/ca.crt"
	if _, err := os.Stat(keyPath); err == nil {
		if _, err := os.Stat(certPath); err == nil {
			// Both files exist, do not overwrite
			return nil
		}
	}

	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	caTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: caCN},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}
	// Write CA key
	keyOut, _ := os.Create(keyPath)
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})
	// Write CA cert
	certOut, _ := os.Create(certPath)
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	return nil
}

// Generate manager key and certificate signed by CA, with SANs
func GenerateManagerCert(caKey *rsa.PrivateKey, caCert *x509.Certificate, managerCN string, sanHosts []net.IP) error {
	managerKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	managerTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: managerCN},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(5 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           sanHosts, // Add IP SANs here
	}
	managerDER, err := x509.CreateCertificate(rand.Reader, &managerTemplate, caCert, &managerKey.PublicKey, caKey)
	if err != nil {
		return err
	}
	// Write manager key
	keyOut, _ := os.Create(config.CertsPath + "/manager.key")
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(managerKey)})
	// Write manager cert
	certOut, _ := os.Create(config.CertsPath + "/manager.crt")
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: managerDER})
	return nil
}

// Helper to load CA key/cert for signing manager certs
func LoadCA() (*rsa.PrivateKey, *x509.Certificate, error) {
	caKeyPEM, err := os.ReadFile(config.CertsPath + "/ca.key")
	if err != nil {
		return nil, nil, err
	}
	caCertPEM, err := os.ReadFile(config.CertsPath + "/ca.crt")
	if err != nil {
		return nil, nil, err
	}
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	caCertBlock, _ := pem.Decode(caCertPEM)
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return caKey, caCert, nil
}

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
