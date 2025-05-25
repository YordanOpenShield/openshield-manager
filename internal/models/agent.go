package models

import (
	"time"

	"github.com/google/uuid"
)

type AgentState string

const (
	AgentStateDisconnected AgentState = "DISCONNECTED" // Agent is registered but not connected
	AgentStateConnected    AgentState = "CONNECTED"    // Agent is registered and connected
)

type Agent struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	DeviceID string     `gorm:"uniqueIndex" json:"device_id"`
	Token    string     `json:"token"`
	LastSeen time.Time  `json:"last_seen"`
	Address  string     `gorm:"column:address" json:"address"`
	State    AgentState `gorm:"default:'DISCONNECTED'" json:"state"`
}

type AgentAddress struct {
	AgentID uuid.UUID `gorm:"type:uuid" json:"agent_id"`
	Address string    `json:"address"`
}

type AgentService struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AgentID   uuid.UUID `gorm:"type:uuid;index" json:"agent_id"`
	Name      string    `json:"name"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"` // e.g., "tcp", "udp"
	Status    string    `json:"status"`   // e.g., "running", "stopped"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
