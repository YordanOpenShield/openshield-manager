package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ToolAction struct {
	Name string   `json:"name"`
	Opts []string `json:"opts"`
}

type Tool struct {
	Name    string       `json:"name"`
	Actions []ToolAction `json:"actions"`
	OS      []string     `json:"os"`
}

type ToolActionExecution struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid;index"`
	ToolName       string     `json:"tool_name" gorm:"type:varchar(255);not null"`
	ToolAction     string     `json:"tool_action" gorm:"type:varchar(255);not null"`
	ToolOptions    []string   `json:"tool_options" gorm:"type:text"` // Will require custom type or serializer for []string
	AgentID        uuid.UUID  `json:"agent_id" gorm:"type:uuid;not null"`
	Status         string     `json:"status" gorm:"default:'PENDING';not null"` // e.g., "PENDING", "RUNNING", "COMPLETED", "FAILED"
	Result         string     `json:"result" gorm:"type:text"`                  // Result of the tool action execution
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`         // Automatically set to current time on creation
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`         // Automatically set to current time on update
}

func (t *ToolActionExecution) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
