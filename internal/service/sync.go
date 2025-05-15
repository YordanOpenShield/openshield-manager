package service

import (
	"log"
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
