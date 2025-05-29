package api

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"openshield-manager/internal/config"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// POST /api/cert/sign
func SignAgentCSR(c *gin.Context) {
	// Check if the request has the required header
	token := c.GetHeader("X-Agent-Token")
	var agent models.Agent
	if err := db.DB.Where("token = ?", token).First(&agent).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent token"})
		return
	}

	// Read CSR from request body
	csrPEM, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSR"})
		return
	}
	block, _ := pem.Decode(csrPEM)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSR PEM"})
		return
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSR"})
		return
	}
	if err := csr.CheckSignature(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSR signature"})
		return
	}

	// Load CA cert and key
	caCertPEM, err := os.ReadFile(config.CertsPath + "/ca.crt")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CA cert"})
		return
	}
	caKeyPEM, err := os.ReadFile(config.CertsPath + "/ca.key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CA key"})
		return
	}
	caBlock, _ := pem.Decode(caCertPEM)
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caBlock == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid CA cert PEM"})
		return
	}
	if caKeyBlock == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid CA key PEM"})
		return
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CA cert"})
		return
	}

	var caKey interface{}
	caKey, err = x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		caKey, err = x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CA key"})
			return
		}
	}

	agentCertTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UTC().UnixNano()),
		Subject:               csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		nil, agentCertTmpl, caCert, csr.PublicKey, caKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign certificate"})
		return
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Respond with agent cert and CA cert
	c.JSON(http.StatusOK, gin.H{
		"agent_cert": string(certPEM),
		"ca_cert":    string(caCertPEM),
	})
}
