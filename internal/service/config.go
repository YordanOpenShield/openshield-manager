package service

import (
	"log"
	"time"

	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"
)

// ConfigSyncMonitor starts a goroutine that syncs configurations for all connected agents every N seconds.
func ConfigSyncMonitor(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var agents []models.Agent
				if err := db.DB.Where("state = ? AND address != ''", "CONNECTED").Find(&agents).Error; err != nil {
					log.Printf("[CONFIG SYNC] Failed to query agents: %v", err)
					continue
				}
				for _, agent := range agents {
					go func(agent models.Agent) {
						if err := managergrpc.SyncConfigs(agent.Address); err != nil {
							log.Printf("[CONFIG SYNC] Failed to sync configs for agent %s: %v", agent.ID, err)
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
