package api

import (
	"net/http"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetJobs(c *gin.Context) {
	orgID, _ := utils.GetOrgID(c)
	var jobs []models.Job
	db.DB.Scopes(db.TenantScope(orgID)).Find(&jobs)
	c.JSON(http.StatusOK, jobs)
}

func GetJobDetails(c *gin.Context) {
	id := c.Param("id")
	orgID, _ := utils.GetOrgID(c)
	var job models.Job
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

type CreateJobRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Target      string `json:"target"`
}

func CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	orgID, _ := utils.GetOrgID(c)

	// Create job in DB
	job := models.Job{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		Type:           models.JobType(req.Type),
		Target:         req.Target,
		OrganizationID: orgID,
	}
	if err := db.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}
