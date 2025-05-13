package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Job struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Command     string    `gorm:"type:text;not null" json:"command"`
}

func (t *Job) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
