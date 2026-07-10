package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"openshield-manager/internal/config"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already in use")
)

// AuthClaims represents the JWT claims for authenticated users.
type AuthClaims struct {
	UserID         string  `json:"user_id"`
	Email          string  `json:"email"`
	Role           string  `json:"role"`
	OrganizationID *string `json:"organization_id,omitempty"`
	jwt.RegisteredClaims
}

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// AuthenticateUser validates email+password and returns a JWT token string.
func AuthenticateUser(email, password string) (string, *models.User, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !CheckPassword(password, user.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := GenerateUserToken(&user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// GenerateUserToken creates a signed JWT for the given user.
func GenerateUserToken(user *models.User) (string, error) {
	secret := config.GlobalConfig.JWT_SECRET
	if secret == "" {
		secret = "change-me-in-production"
	}

	var orgIDStr *string
	if user.OrganizationID != nil {
		s := user.OrganizationID.String()
		orgIDStr = &s
	}

	claims := AuthClaims{
		UserID:         user.ID.String(),
		Email:          user.Email,
		Role:           string(user.Role),
		OrganizationID: orgIDStr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateUserToken parses and validates a JWT token string, returning the claims.
func ValidateUserToken(tokenString string) (*AuthClaims, error) {
	secret := config.GlobalConfig.JWT_SECRET
	if secret == "" {
		secret = "change-me-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RegisterUser creates a new user with a hashed password.
func RegisterUser(email, password, name string, role models.UserRole, orgID *uuid.UUID) (*models.User, error) {
	// Check if email already exists
	var existing models.User
	if err := db.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:          email,
		PasswordHash:   hashedPassword,
		Name:           name,
		Role:           role,
		OrganizationID: orgID,
	}

	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// RevokeUserTokens invalidates all tokens for a user by changing their password hash.
// In a production system, you'd use a token blacklist. For now, we simply
// rely on token expiry. A full revocation requires a deny list or short TTL.
func RevokeUserTokens(userID uuid.UUID) error {
	// Touch the user's updated_at to invalidate tokens if we implement a cache later.
	// For now this is a placeholder — tokens remain valid until expiry.
	return db.DB.Model(&models.User{}).Where("id = ?", userID).Update("updated_at", time.Now()).Error
}
