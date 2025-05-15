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
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	DeviceID string    `gorm:"uniqueIndex"`
	Token    string
	LastSeen time.Time
	Address  string
	State    AgentState `gorm:"default:'Unregistered'"`
}
