package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	JobID     uuid.UUID  `gorm:"not null" json:"job_id"`
	AgentID   uuid.UUID  `gorm:"not null" json:"agent_id"`
	Status    TaskStatus `gorm:"default:'pending'" json:"status"`
	Result    string     `gorm:"type:text" json:"result"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
