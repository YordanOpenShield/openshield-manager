package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BulkOperationType represents the type of bulk operation
type BulkOperationType string

const (
	BulkOpTaskAssign  BulkOperationType = "TASK_ASSIGN"
	BulkOpQueryRun    BulkOperationType = "QUERY_RUN"
	BulkOpToolExecute BulkOperationType = "TOOL_EXECUTE"
	BulkOpUpdate      BulkOperationType = "UPDATE"
	BulkOpRestart     BulkOperationType = "RESTART"
)

// BulkOperationStatus represents the status of a bulk operation
type BulkOperationStatus string

const (
	BulkOpPending   BulkOperationStatus = "PENDING"
	BulkOpRunning   BulkOperationStatus = "RUNNING"
	BulkOpCompleted BulkOperationStatus = "COMPLETED"
	BulkOpPartial   BulkOperationStatus = "PARTIAL"
	BulkOpFailed    BulkOperationStatus = "FAILED"
	BulkOpCancelled BulkOperationStatus = "CANCELLED"
)

// BulkOperation represents a bulk operation on multiple agents
type BulkOperation struct {
	ID             uuid.UUID           `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID *uuid.UUID          `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Type           BulkOperationType   `json:"type"`
	Target         BulkTarget          `gorm:"type:jsonb" json:"target"`
	Payload        interface{}         `gorm:"type:jsonb" json:"payload"`
	Status         BulkOperationStatus `gorm:"default:'PENDING'" json:"status"`
	Progress       BulkProgress        `gorm:"type:jsonb" json:"progress"`
	Results        []BulkResult        `gorm:"type:jsonb" json:"results"`
	CreatedAt      time.Time           `json:"created_at"`
	CompletedAt    *time.Time          `json:"completed_at,omitempty"`
}

// BulkTarget defines which agents to target
type BulkTarget struct {
	AgentIDs     []string `json:"agent_ids,omitempty"`
	GroupID      string   `json:"group_id,omitempty"`
	AllConnected bool     `json:"all_connected,omitempty"`
}

// BulkProgress tracks operation progress
type BulkProgress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Pending   int `json:"pending"`
}

// BulkResult represents the result for a single agent
type BulkResult struct {
	AgentID    string      `json:"agent_id"`
	Status     string      `json:"status"` // SUCCESS, ERROR, TIMEOUT
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	ExecutedAt time.Time   `json:"executed_at"`
}

// BeforeCreate hook to generate UUID
func (b *BulkOperation) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
