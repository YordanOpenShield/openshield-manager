package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organization represents a customer organization in the multi-tenant system.
type Organization struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (o *Organization) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}
