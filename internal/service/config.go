package service

import (
	"log"
	"time"

	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"
)

// ConfigSyncMonitor starts a goroutine that syncs configurations for all connected agents every N seconds.
// It iterates per-organization to ensure proper org scoping.
func ConfigSyncMonitor(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get all organizations
				var orgs []models.Organization
				if err := db.DB.Find(&orgs).Error; err != nil {
					log.Printf("[CONFIG SYNC] Failed to query organizations: %v", err)
					continue
				}
				// Sync agents per org
				for _, org := range orgs {
					orgID := org.ID
					var agents []models.Agent
					if err := db.DB.Scopes(db.TenantScope(&orgID)).Where("state = ? AND address != ''", "CONNECTED").Find(&agents).Error; err != nil {
						log.Printf("[CONFIG SYNC] Failed to query agents for org %s: %v", orgID, err)
						continue
					}
					for _, agent := range agents {
						go func(agent models.Agent) {
							if err := managergrpc.SyncConfigs(agent.Address); err != nil {
								log.Printf("[CONFIG SYNC] Failed to sync configs for agent %s (org %s): %v", agent.ID, orgID, err)
							}
						}(agent)
					}
				}
				// Also handle legacy agents without an org
				var legacyAgents []models.Agent
				if err := db.DB.Where("organization_id IS NULL AND state = ? AND address != ''", "CONNECTED").Find(&legacyAgents).Error; err != nil {
					log.Printf("[CONFIG SYNC] Failed to query legacy agents: %v", err)
					continue
				}
				for _, agent := range legacyAgents {
					go func(agent models.Agent) {
						if err := managergrpc.SyncConfigs(agent.Address); err != nil {
							log.Printf("[CONFIG SYNC] Failed to sync configs for legacy agent %s: %v", agent.ID, err)
						}
					}(agent)
				}
			case <-stopCh:
				log.Println("[CONFIG SYNC] Global configs sync monitor stopped.")
				return
			}
		}
	}()
}
