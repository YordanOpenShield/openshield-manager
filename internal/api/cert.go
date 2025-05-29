package api

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// POST /api/cert/sign
// TODO: Export this function to the utils package (cert.go)
type SignAgentCSRRequest struct {
	SANHosts []string `json:"sanHosts" binding:"required"`
	CSR      string   `json:"csr" binding:"required"` // PEM as string
}

func SignAgentCSR(c *gin.Context) {
	// Check if the request has the required header
	token := c.GetHeader("X-Agent-Token")
	var agent models.Agent
	if err := db.DB.Where("token = ?", token).First(&agent).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent token"})
		return
	}

	var req SignAgentCSRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert SANHosts to IPs and DNS names
	sanHosts := make([]net.IP, 0, len(req.SANHosts))
	for _, addr := range req.SANHosts {
		ip := net.ParseIP(addr)
		if ip != nil {
			sanHosts = append(sanHosts, ip)
		}
	}

	// Read CSR from request body
	csrPEM := []byte(req.CSR)
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
	caKey, caCert, err := utils.LoadCA()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load CA certificate"})
		return
	}

	agentCertTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UTC().UnixNano()),
		Subject:               csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           sanHosts, // Add IP SANs here
	}

	certDER, err := x509.CreateCertificate(
		nil, agentCertTmpl, caCert, csr.PublicKey, caKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign certificate"})
		return
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Encode CA cert to PEM
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})

	// Respond with agent cert and CA cert
	c.JSON(http.StatusOK, gin.H{
		"agent_cert": string(certPEM),
		"ca_cert":    string(caCertPEM),
	})
}
