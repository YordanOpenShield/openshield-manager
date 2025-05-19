package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobType string

const (
	JobTypeCommand JobType = "COMMAND"
	JobTypeScript  JobType = "SCRIPT"
)

type Job struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Type        JobType   `gorm:"type:varchar(16);not null" json:"type"`
	Target      string    `gorm:"type:text;not null" json:"target"`
}

func (t *Job) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
