package models

import (
	"github.com/google/uuid"
)

type Job struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	AgentID uuid.UUID
	Tool    string
	Action  string
	Status  string
}
