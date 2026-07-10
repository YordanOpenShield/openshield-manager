package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistrationTokenStatus represents the status of a registration token.
type RegistrationTokenStatus string

const (
	RegTokenActive  RegistrationTokenStatus = "ACTIVE"
	RegTokenUsed    RegistrationTokenStatus = "USED"
	RegTokenExpired RegistrationTokenStatus = "EXPIRED"
	RegTokenRevoked RegistrationTokenStatus = "REVOKED"
)

// RegistrationToken is a one-time token that binds an agent to an organization during registration.
type RegistrationToken struct {
	ID             uuid.UUID               `gorm:"type:uuid;primaryKey" json:"id"`
	Token          string                  `gorm:"uniqueIndex;not null" json:"token"`
	OrganizationID uuid.UUID               `gorm:"type:uuid;not null;index" json:"organization_id"`
	Organization   Organization            `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Status         RegistrationTokenStatus `gorm:"type:varchar(16);default:'ACTIVE'" json:"status"`
	MaxUses        int                     `gorm:"default:1" json:"max_uses"`
	UseCount       int                     `gorm:"default:0" json:"use_count"`
	CreatedBy      uuid.UUID               `gorm:"type:uuid" json:"created_by"`
	CreatedAt      time.Time               `json:"created_at"`
	ExpiresAt      *time.Time              `json:"expires_at,omitempty"`
}

func (t *RegistrationToken) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
