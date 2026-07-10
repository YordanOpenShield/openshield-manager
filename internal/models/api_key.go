package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApiKey represents a programmatic API key scoped to an organization.
type ApiKey struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	KeyHash        string       `gorm:"not null" json:"-"`
	Name           string       `gorm:"not null" json:"name"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	ExpiresAt      *time.Time   `json:"expires_at,omitempty"`
}

func (k *ApiKey) BeforeCreate(tx *gorm.DB) (err error) {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return
}
