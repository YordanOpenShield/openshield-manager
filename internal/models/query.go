package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Query represents a saved query that can be run on agents
type Query struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID *uuid.UUID `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Name           string     `gorm:"not null" json:"name"`
	Description    string     `json:"description"`
	SQL            string     `gorm:"type:text;not null" json:"sql"`
	Platform       string     `json:"platform"` // linux, windows, darwin, or empty for all
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (q *Query) BeforeCreate(tx *gorm.DB) (err error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return
}

// QueryStatus represents the status of a query execution
type QueryStatus string

const (
	QueryStatusPending   QueryStatus = "PENDING"
	QueryStatusRunning   QueryStatus = "RUNNING"
	QueryStatusCompleted QueryStatus = "COMPLETED"
	QueryStatusFailed    QueryStatus = "FAILED"
	QueryStatusCancelled QueryStatus = "CANCELLED"
)

// QueryExecution tracks a query execution on one or more agents
type QueryExecution struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID *uuid.UUID  `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	QueryID        uuid.UUID   `gorm:"type:uuid;not null" json:"query_id"`
	Query          Query       `gorm:"foreignKey:QueryID" json:"query,omitempty"`
	Status         QueryStatus `gorm:"default:'PENDING'" json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty"`
}

func (q *QueryExecution) BeforeCreate(tx *gorm.DB) (err error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return
}

// QueryExecutionResult stores results for each agent in a query execution
type QueryExecutionResult struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID *uuid.UUID  `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	ExecutionID    uuid.UUID   `gorm:"type:uuid;not null" json:"execution_id"`
	AgentID        uuid.UUID   `gorm:"type:uuid;not null" json:"agent_id"`
	Agent          Agent       `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
	Status         QueryStatus `gorm:"default:'PENDING'" json:"status"`
	ResultJSON     string      `gorm:"type:text" json:"result_json"`
	Error          string      `json:"error"`
	StartedAt      *time.Time  `json:"started_at,omitempty"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty"`
}

func (q *QueryExecutionResult) BeforeCreate(tx *gorm.DB) (err error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return
}

// RunQueryRequest represents a request to run a query on agents
type RunQueryRequest struct {
	QueryID  string   `json:"query_id" binding:"required"`
	AgentIDs []string `json:"agent_ids" binding:"required"` // empty means all agents
}

// CreateQueryRequest represents a request to create a new query
type CreateQueryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SQL         string `json:"sql" binding:"required"`
	Platform    string `json:"platform"`
}
