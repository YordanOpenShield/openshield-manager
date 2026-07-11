package service

import (
	"log"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
)

// AgentLastSeenMonitor starts a goroutine that checks agent last seen timestamps every N seconds.
// It iterates per-organization to ensure proper org scoping.
func AgentLastSeenMonitor(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				// Get all organizations
				var orgs []models.Organization
				if err := db.DB.Find(&orgs).Error; err != nil {
					log.Printf("[AGENT MONITOR] Failed to query organizations: %v", err)
					continue
				}
				// Process agents per org
				for _, org := range orgs {
					orgID := org.ID
					var agents []models.Agent
					if err := db.DB.Scopes(db.TenantScope(&orgID)).Find(&agents).Error; err != nil {
						log.Printf("[AGENT MONITOR] Failed to query agents for org %s: %v", orgID, err)
						continue
					}
					for _, agent := range agents {
						if agent.LastSeen.Before(now.Add(-30*time.Second)) && agent.State != "DISCONNECTED" {
							if err := db.DB.Scopes(db.TenantScope(&orgID)).Model(&models.Agent{}).
								Where("id = ?", agent.ID).
								Update("state", "DISCONNECTED").Error; err != nil {
								log.Printf("[AGENT MONITOR] Failed to mark agent %s as disconnected: %v", agent.ID, err)
							} else {
								log.Printf("[AGENT MONITOR] Agent %s (org %s) marked as DISCONNECTED due to inactivity", agent.ID, orgID)
							}
						}
					}
				}
				// Also handle legacy agents without an org
				var legacyAgents []models.Agent
				if err := db.DB.Where("organization_id IS NULL").Find(&legacyAgents).Error; err != nil {
					log.Printf("[AGENT MONITOR] Failed to query legacy agents: %v", err)
					continue
				}
				for _, agent := range legacyAgents {
					if agent.LastSeen.Before(now.Add(-30*time.Second)) && agent.State != "DISCONNECTED" {
						if err := db.DB.Where("organization_id IS NULL AND id = ?", agent.ID).Model(&models.Agent{}).
							Update("state", "DISCONNECTED").Error; err != nil {
							log.Printf("[AGENT MONITOR] Failed to mark legacy agent %s as disconnected: %v", agent.ID, err)
						} else {
							log.Printf("[AGENT MONITOR] Legacy agent %s marked as DISCONNECTED due to inactivity", agent.ID)
						}
					}
				}
			case <-stopCh:
				log.Println("[AGENT MONITOR] Last seen monitor stopped.")
				return
			}
		}
	}()
}
