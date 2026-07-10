package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openshield-manager/internal/middleware"
	"openshield-manager/internal/service"
)

// ListOrganizations handles GET /api/organizations
func ListOrganizations(c *gin.Context) {
	orgs, err := service.ListOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list organizations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  orgs,
		"error": 0,
	})
}

// GetOrganization handles GET /api/organizations/:id
func GetOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := service.GetOrganizationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  org,
		"error": 0,
	})
}

// CreateOrganization handles POST /api/organizations
func CreateOrganization(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	org, err := service.CreateOrganization(req.Name, req.Slug)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":  org,
		"error": 0,
	})
}

// UpdateOrganization handles PUT /api/organizations/:id
func UpdateOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	org, err := service.UpdateOrganization(id, req.Name, req.Slug)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  org,
		"error": 0,
	})
}

// DeleteOrganization handles DELETE /api/organizations/:id
func DeleteOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	if err := service.DeleteOrganization(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  gin.H{"message": "Organization deleted"},
		"error": 0,
	})
}

// RegisterOrgRoutes adds organization management endpoints to the given API group.
func RegisterOrgRoutes(apiGroup *gin.RouterGroup) {
	orgs := apiGroup.Group("/organizations")
	{
		orgs.GET("", middleware.AuthMiddleware(), middleware.RequireRole(middleware.RoleAdmin), ListOrganizations)
		orgs.GET("/:id", middleware.AuthMiddleware(), middleware.RequireRole(middleware.RoleAdmin), GetOrganization)
		orgs.POST("", middleware.AuthMiddleware(), middleware.RequireRole(middleware.RoleSuperAdmin), CreateOrganization)
		orgs.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRole(middleware.RoleSuperAdmin), UpdateOrganization)
		orgs.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRole(middleware.RoleSuperAdmin), DeleteOrganization)
	}
}
