package models

import (
	"time"

	"github.com/google/uuid"
)

type AgentState string

const (
	AgentStateDisconnected AgentState = "DISCONNECTED" // Agent is registered but not connected
	AgentStateConnected    AgentState = "CONNECTED"    // Agent is registered and connected
	AgentStateUnregistered AgentState = "UNREGISTERED" // Agent is not registered
)

type Agent struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	DeviceID string     `gorm:"uniqueIndex" json:"device_id"`
	Token    string     `json:"token"`
	LastSeen time.Time  `json:"last_seen"`
	Address  string     `json:"address"`
	State    AgentState `gorm:"default:'Unregistered'" json:"state"`
}

type AgentAddress struct {
	AgentID uuid.UUID `gorm:"type:uuid" json:"agent_id"`
	Address string    `json:"ip"`
}
