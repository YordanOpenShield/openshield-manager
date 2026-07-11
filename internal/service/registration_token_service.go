package service

import (
	"errors"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRegTokenNotFound = errors.New("registration token not found")
)

// CreateRegistrationToken creates a new registration token for an organization.
func CreateRegistrationToken(orgID uuid.UUID, maxUses int, expiresIn *time.Duration, createdBy uuid.UUID) (*models.RegistrationToken, error) {
	tokenStr := "osh_reg_" + uuid.New().String()
	token := models.RegistrationToken{
		Token:          tokenStr,
		OrganizationID: orgID,
		Status:         models.RegTokenActive,
		MaxUses:        maxUses,
		UseCount:       0,
		CreatedBy:      createdBy,
	}

	if expiresIn != nil {
		exp := time.Now().Add(*expiresIn)
		token.ExpiresAt = &exp
	}

	if err := db.DB.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

// ListRegistrationTokens returns registration tokens, optionally scoped to an org.
// Pass nil orgID for super admins (returns all), or a specific orgID for org-scoped listing.
func ListRegistrationTokens(orgID *uuid.UUID) ([]models.RegistrationToken, error) {
	var tokens []models.RegistrationToken
	query := db.DB.Preload("Organization").Order("created_at DESC")
	if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	}
	if err := query.Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetRegistrationTokenByID returns a registration token by its ID.
func GetRegistrationTokenByID(id uuid.UUID) (*models.RegistrationToken, error) {
	var token models.RegistrationToken
	if err := db.DB.Preload("Organization").Where("id = ?", id).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

// RevokeRegistrationToken revokes a registration token (sets status to REVOKED).
func RevokeRegistrationToken(id uuid.UUID) error {
	result := db.DB.Model(&models.RegistrationToken{}).
		Where("id = ?", id).
		Where("status IN ?", []models.RegistrationTokenStatus{models.RegTokenActive, models.RegTokenExpired}).
		Update("status", models.RegTokenRevoked)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRegTokenNotFound
	}
	return nil
}
