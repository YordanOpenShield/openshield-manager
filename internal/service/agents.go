package service

import (
	"log"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
)

// StartAgentLastSeenMonitor starts a goroutine that checks agent last seen timestamps every 30 seconds
func AgentLastSeenMonitor(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				var agents []models.Agent
				if err := db.DB.Find(&agents).Error; err != nil {
					log.Printf("[AGENT MONITOR] Failed to query agents: %v", err)
					continue
				}
				for _, agent := range agents {
					if agent.LastSeen.Before(now.Add(-30*time.Second)) && agent.State != "DISCONNECTED" {
						if err := db.DB.Model(&models.Agent{}).
							Where("id = ?", agent.ID).
							Update("state", "DISCONNECTED").Error; err != nil {
							log.Printf("[AGENT MONITOR] Failed to mark agent %s as disconnected: %v", agent.ID, err)
						} else {
							log.Printf("[AGENT MONITOR] Agent %s marked as DISCONNECTED due to inactivity", agent.ID)
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
