package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system.
type UserRole string

const (
	UserRoleSuperAdmin  UserRole = "super_admin"
	UserRoleOrgAdmin    UserRole = "org_admin"
	UserRoleOrgViewer   UserRole = "org_viewer"
	UserRoleOrgOperator UserRole = "org_operator"
)

// User represents an authenticated user (dashboard/admin).
type User struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Email          string        `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string        `gorm:"not null" json:"-"`
	Name           string        `json:"name"`
	Role           UserRole      `gorm:"type:varchar(32);default:'org_viewer'" json:"role"`
	OrganizationID *uuid.UUID    `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	AllowRunAs     bool          `gorm:"default:false" json:"allow_run_as"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
