package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantScope returns a GORM scope that filters queries by organization_id.
// Pass nil for the orgID to skip scoping (for super admins or cross-org queries).
// Usage: db.Scopes(TenantScope(&orgID)).Find(&agents)
func TenantScope(orgID *uuid.UUID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID == nil {
			return db
		}
		return db.Where("organization_id = ?", *orgID)
	}
}

// TenantScopeWithFallback returns a GORM scope that filters by organization_id.
// If orgID is nil, it filters where organization_id IS NULL (for unassigned records).
func TenantScopeWithFallback(orgID *uuid.UUID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID == nil {
			return db.Where("organization_id IS NULL")
		}
		return db.Where("organization_id = ?", *orgID)
	}
}

// OrgScope is an alias for TenantScope for shorter usage in handlers.
func OrgScope(orgID *uuid.UUID) func(*gorm.DB) *gorm.DB {
	return TenantScope(orgID)
}
