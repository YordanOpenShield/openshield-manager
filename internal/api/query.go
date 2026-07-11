package api

import (
	"net/http"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetQueries returns all saved queries scoped to the user's organization
func GetQueries(c *gin.Context) {
	orgID, _ := utils.GetOrgID(c)
	var queries []models.Query
	if err := db.DB.Scopes(db.TenantScope(orgID)).Find(&queries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queries"})
		return
	}
	c.JSON(http.StatusOK, queries)
}

// GetQuery returns a specific query by ID, scoped to the user's organization
func GetQuery(c *gin.Context) {
	id := c.Param("id")
	orgID, _ := utils.GetOrgID(c)
	var query models.Query
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).First(&query).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		return
	}
	c.JSON(http.StatusOK, query)
}

// CreateQuery creates a new saved query scoped to the user's organization
func CreateQuery(c *gin.Context) {
	var req models.CreateQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	orgID, _ := utils.GetOrgID(c)

	query := models.Query{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		SQL:            req.SQL,
		Platform:       req.Platform,
		OrganizationID: orgID,
	}

	if err := db.DB.Create(&query).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create query: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, query)
}

// UpdateQuery updates an existing query, scoped to the user's organization
func UpdateQuery(c *gin.Context) {
	id := c.Param("id")
	orgID, _ := utils.GetOrgID(c)
	var query models.Query
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).First(&query).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		return
	}

	var req models.CreateQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	query.Name = req.Name
	query.Description = req.Description
	query.SQL = req.SQL
	query.Platform = req.Platform

	if err := db.DB.Save(&query).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update query: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, query)
}

// DeleteQuery deletes a query, scoped to the user's organization
func DeleteQuery(c *gin.Context) {
	id := c.Param("id")
	orgID, _ := utils.GetOrgID(c)

	// Find all executions for this query
	var executions []models.QueryExecution
	db.DB.Scopes(db.TenantScope(orgID)).Where("query_id = ?", id).Find(&executions)

	for _, exec := range executions {
		// Delete execution results first
		db.DB.Scopes(db.TenantScope(orgID)).Where("execution_id = ?", exec.ID).Delete(&models.QueryExecutionResult{})
		// Then delete the execution
		db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", exec.ID).Delete(&models.QueryExecution{})
	}

	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).Delete(&models.Query{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete query: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Query deleted successfully"})
}

// GetQueryExecutions returns all query executions scoped to the user's organization
func GetQueryExecutions(c *gin.Context) {
	orgID, _ := utils.GetOrgID(c)
	var executions []models.QueryExecution
	if err := db.DB.Scopes(db.TenantScope(orgID)).Preload("Query").Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executions"})
		return
	}
	c.JSON(http.StatusOK, executions)
}

// GetQueryExecution returns a specific execution with results, scoped to the user's organization
func GetQueryExecution(c *gin.Context) {
	id := c.Param("id")
	orgID, _ := utils.GetOrgID(c)
	var execution models.QueryExecution
	if err := db.DB.Scopes(db.TenantScope(orgID)).Preload("Query").Where("id = ?", id).First(&execution).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}

	var results []models.QueryExecutionResult
	if err := db.DB.Scopes(db.TenantScope(orgID)).Preload("Agent").Where("execution_id = ?", execution.ID).Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution": execution,
		"results":   results,
	})
}
