package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentState string

const (
	AgentStateDisconnected AgentState = "DISCONNECTED" // Agent is registered but not connected
	AgentStateConnected    AgentState = "CONNECTED"    // Agent is registered and connected
)

type Agent struct {
	ID       uuid.UUID              `gorm:"type:uuid;primaryKey" json:"id"`
	DeviceID string                 `gorm:"uniqueIndex" json:"device_id"`
	Token    string                 `json:"token"`
	LastSeen time.Time              `json:"last_seen"`
	Address  string                 `gorm:"column:address" json:"address"`
	State    AgentState             `gorm:"default:'DISCONNECTED'" json:"state"`
	Metadata map[string]interface{} `gorm:"type:jsonb" json:"metadata"`
}

// AgentGroup represents a group of agents (static or dynamic)
type AgentGroup struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Tags        []string      `gorm:"type:jsonb;serializer:json" json:"tags"`
	IsDynamic   bool           `gorm:"default:false" json:"is_dynamic"`
	Criteria    *AgentCriteria `gorm:"type:jsonb" json:"criteria,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AgentCriteria for dynamic groups
type AgentCriteria struct {
	OS             []string          `json:"os,omitempty"`
	Version        string            `json:"version,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	LastSeenWithin int               `json:"last_seen_within,omitempty"` // seconds
	State          []string          `json:"state,omitempty"`
}

// GroupMembership links agents to groups (for static groups)
type GroupMembership struct {
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	AgentID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate hook for AgentGroup
func (g *AgentGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
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
