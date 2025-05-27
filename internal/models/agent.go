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
	AgentID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"agent_id"`
	Name      string    `gorm:"primaryKey" json:"name"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
