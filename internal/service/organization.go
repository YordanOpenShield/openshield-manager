package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshield-manager/internal/config"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
)

var (
	ErrOrgNotFound   = errors.New("organization not found")
	ErrOrgSlugExists = errors.New("organization slug already exists")
	ErrOrgNameExists = errors.New("organization name already exists")
)

// CreateOrganization creates a new organization with a generated slug.
func CreateOrganization(name, slug string) (*models.Organization, error) {
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	// Check slug uniqueness
	var existing models.Organization
	if err := db.DB.Where("slug = ?", slug).First(&existing).Error; err == nil {
		return nil, ErrOrgSlugExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	org := &models.Organization{
		Name: name,
		Slug: slug,
	}

	if err := db.DB.Create(org).Error; err != nil {
		return nil, err
	}

	return org, nil
}

// GetOrganizationByID retrieves an organization by its UUID.
func GetOrganizationByID(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	if err := db.DB.First(&org, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// ListOrganizations returns all organizations.
func ListOrganizations() ([]models.Organization, error) {
	var orgs []models.Organization
	if err := db.DB.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// UpdateOrganization modifies an existing organization.
func UpdateOrganization(id uuid.UUID, name, slug string) (*models.Organization, error) {
	org, err := GetOrganizationByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		org.Name = name
	}
	if slug != "" {
		// Check slug uniqueness (exclude self)
		var existing models.Organization
		if err := db.DB.Where("slug = ? AND id != ?", slug, id).First(&existing).Error; err == nil {
			return nil, ErrOrgSlugExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		org.Slug = slug
	}

	if err := db.DB.Save(org).Error; err != nil {
		return nil, err
	}

	return org, nil
}

// DeleteOrganization removes an organization by ID.
func DeleteOrganization(id uuid.UUID) error {
	result := db.DB.Delete(&models.Organization{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return ErrOrgNotFound
	}
	return result.Error
}

// SeedDefaultOrg creates a default organization and super admin if no users exist.
func SeedDefaultOrg() error {
	var userCount int64
	db.DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return nil // Already seeded
	}

	// Create a default organization
	_, err := CreateOrganization("Default Organization", "default")
	if err != nil && !errors.Is(err, ErrOrgSlugExists) {
		return err
	}
	if err == nil {
		// Use config values with sensible defaults
		email := config.GlobalConfig.ADMIN_EMAIL
		if email == "" {
			email = "admin@openshield.local"
		}
		password := config.GlobalConfig.ADMIN_PASSWORD
		if password == "" {
			password = "admin"
		}
		name := config.GlobalConfig.ADMIN_NAME
		if name == "" {
			name = "Super Admin"
		}

		// Create super admin user
		_, err = RegisterUser(
			email,
			password,
			name,
			models.UserRoleSuperAdmin,
			nil, // no org for super admin
		)
		if err != nil {
			return err
		}
	}

	return nil
}
