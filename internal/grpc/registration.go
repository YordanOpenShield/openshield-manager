package managergrpc

import (
	"errors"
	"log"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRegTokenNotFound    = errors.New("registration token not found")
	ErrRegTokenInvalid     = errors.New("registration token is invalid or expired")
	ErrRegTokenExhausted   = errors.New("registration token has been fully used")
	ErrRegTokenOrgNotFound = errors.New("organization for registration token not found")
)

// ValidateRegistrationToken checks if a registration token is valid and returns the associated organization ID.
// It also increments the use count and marks as USED if max_uses reached.
func ValidateRegistrationToken(tokenStr string) (*uuid.UUID, error) {
	if tokenStr == "" {
		return nil, ErrRegTokenNotFound
	}

	var token models.RegistrationToken
	if err := db.DB.Where("token = ?", tokenStr).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegTokenNotFound
		}
		return nil, err
	}

	// Check status
	if token.Status != models.RegTokenActive {
		log.Printf("[REG TOKEN] Token %s is %s", tokenStr[:8]+"...", token.Status)
		return nil, ErrRegTokenInvalid
	}

	// Check expiration
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		log.Printf("[REG TOKEN] Token %s expired at %s", tokenStr[:8]+"...", token.ExpiresAt)
		// Auto-expire
		db.DB.Model(&token).Update("status", models.RegTokenExpired)
		return nil, ErrRegTokenInvalid
	}

	// Check max uses
	if token.UseCount >= token.MaxUses {
		log.Printf("[REG TOKEN] Token %s exhausted (used %d/%d)", tokenStr[:8]+"...", token.UseCount, token.MaxUses)
		// Auto-mark as used
		db.DB.Model(&token).Update("status", models.RegTokenUsed)
		return nil, ErrRegTokenExhausted
	}

	// Increment use count
	updates := map[string]interface{}{
		"use_count": token.UseCount + 1,
	}
	if token.UseCount+1 >= token.MaxUses {
		updates["status"] = models.RegTokenUsed
	}

	if err := db.DB.Model(&token).Updates(updates).Error; err != nil {
		log.Printf("[REG TOKEN] Failed to update token %s: %v", tokenStr[:8]+"...", err)
		return nil, err
	}

	// Verify the organization exists
	var org models.Organization
	if err := db.DB.Where("id = ?", token.OrganizationID).First(&org).Error; err != nil {
		return nil, ErrRegTokenOrgNotFound
	}

	log.Printf("[REG TOKEN] Token %s used (use %d/%d) for org %s",
		tokenStr[:8]+"...", token.UseCount+1, token.MaxUses, org.Name)

	return &token.OrganizationID, nil
}
