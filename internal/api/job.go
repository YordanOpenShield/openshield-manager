package api

import (
	"net/http"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAvailableJobs(c *gin.Context) {
	var jobs []models.Job
	db.DB.Find(&jobs)
	c.JSON(http.StatusOK, jobs)
}

type CreateJobRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

func CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create job in DB
	job := models.Job{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Command:     req.Command,
	}
	if err := db.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}
