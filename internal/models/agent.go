package models

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Token       string
	LastSeen    time.Time
	Environment string
}
