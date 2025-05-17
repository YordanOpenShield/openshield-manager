package service

import (
	"context"
	"log"
	"strings"
	"time"

	"openshield-manager/internal/db"
	grpcclient "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"
)

// StartGlobalScriptSyncMonitor starts a goroutine that syncs scripts for all connected agents every N seconds.
func StartGlobalScriptSyncMonitor(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var agents []models.Agent
				if err := db.DB.Where("state = ? AND address != ''", "CONNECTED").Find(&agents).Error; err != nil {
					log.Printf("[SCRIPT SYNC] Failed to query agents: %v", err)
					continue
				}
				for _, agent := range agents {
					go func(agent models.Agent) {
						if err := grpcclient.SyncScripts(agent.Address); err != nil {
							log.Printf("[SCRIPT SYNC] Failed to sync scripts for agent %s: %v", agent.ID, err)
						}
					}(agent)
				}
			case <-stopCh:
				log.Println("[SCRIPT SYNC] Global script sync monitor stopped.")
				return
			}
		}
	}()
}

func StartAgentHeartbeatMonitor(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Query all agents from DB
				var agents []models.Agent
				if err := db.DB.Find(&agents).Error; err != nil {
					log.Printf("[HEARTBEAT] Failed to query agents: %v", err)
					continue
				}
				for _, agent := range agents {
					go func(agent models.Agent) {
						// Query addresses for this agent from the AgentAddress table
						var addresses []models.AgentAddress
						if err := db.DB.Where("agent_id = ?", agent.ID).Find(&addresses).Error; err != nil {
							log.Printf("[HEARTBEAT SYNC] Failed to query addresses for agent %s: %v", agent.ID, err)
							return
						}
						connected := false
						for _, addrObj := range addresses {
							addr := strings.TrimSpace(addrObj.Address)
							if addr == "" {
								continue
							}
							client, err := grpcclient.NewAgentClient(addr)
							if err != nil {
								log.Printf("[HEARTBEAT SYNC] Could not create client for agent %s at %s: %v", agent.ID, addr, err)
								continue
							}
							ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
							defer cancel()
							ok, err := client.Heartbeat(ctx, agent.ID.String())
							if err == nil && ok {
								connected = true
								// Update the agent's Address field in the database to the working address
								if updateErr := db.DB.Model(&models.Agent{}).
									Where("id = ?", agent.ID).
									Update("address", addr).Error; updateErr != nil {
									log.Printf("[HEARTBEAT SYNC] Failed to update agent %s address: %v", agent.ID, updateErr)
								}
								break
							} else {
								log.Printf("[HEARTBEAT SYNC] Agent %s missed heartbeat at %s: %v", agent.ID, addr, err)
							}
						}
						if !connected {
							// Mark as disconnected in DB
							if updateErr := db.DB.Model(&models.Agent{}).
								Where("id = ?", agent.ID).
								Update("state", "DISCONNECTED").Error; updateErr != nil {
								log.Printf("[HEARTBEAT SYNC] Failed to update agent %s state: %v", agent.ID, updateErr)
							}
						}
					}(agent)
				}
			case <-stopCh:
				log.Println("[HEARTBEAT] Heartbeat monitor stopped.")
				return
			}
		}
	}()
}
