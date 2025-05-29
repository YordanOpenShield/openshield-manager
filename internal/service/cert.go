package service

import (
	"net"
	"openshield-manager/internal/utils"
)

func CreateCertificates() error {
	// CA
	err := utils.GenerateCA("OpenShieldCA")
	if err != nil {
		return err
	}

	// Load CA cert and key
	caKey, caCert, err := utils.LoadCA()
	if err != nil {
		return err
	}

	// Get all possible addresses for the manager
	addresses, err := utils.GetAllLocalAddresses()
	ips := make([]net.IP, 0, len(addresses))
	for _, addr := range addresses {
		ip := net.ParseIP(addr)
		if ip != nil {
			ips = append(ips, ip)
		}
	}
	// Generate manager certificate
	err = utils.GenerateManagerCert(caKey, caCert, "OpenShieldManager", ips)
	if err != nil {
		return err
	}

	return nil
}
