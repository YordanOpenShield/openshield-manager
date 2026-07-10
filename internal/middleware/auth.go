package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"openshield-manager/internal/service"
	"openshield-manager/internal/utils"
)

// AuthMiddleware validates the JWT Bearer token from the Authorization header.
// It extracts user info and sets it in the Gin context for downstream handlers.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format, expected 'Bearer <token>'"})
			return
		}

		tokenString := parts[1]
		claims, err := service.ValidateUserToken(tokenString)
		if err != nil {
			log.Printf("[AUTH] Token validation failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set user info in context
		c.Set(utils.ContextKeyUserID, claims.UserID)
		c.Set(utils.ContextKeyUserEmail, claims.Email)
		c.Set(utils.ContextKeyUserRole, claims.Role)
		c.Set(utils.ContextKeyOrgID, claims.OrganizationID)

		c.Next()
	}
}

// BasicAuthMiddleware validates Basic Auth credentials and returns a JWT.
// This is used only on the login endpoint, matching Wazuh's pattern.
func BasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		email, password, ok := c.Request.BasicAuth()
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid basic auth"})
			return
		}

		// Attempt authentication
		token, user, err := service.AuthenticateUser(email, password)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Set user in context for the handler to use
		c.Set(utils.ContextKeyUserID, user.ID.String())
		c.Set(utils.ContextKeyUserEmail, user.Email)
		c.Set(utils.ContextKeyUserRole, string(user.Role))
		if user.OrganizationID != nil {
			s := user.OrganizationID.String()
			c.Set(utils.ContextKeyOrgID, &s)
		}

		// Store the token and user in the request so the handler can return them
		c.Set("auth_token", token)
		c.Set("auth_user", user)

		c.Next()
	}
}
