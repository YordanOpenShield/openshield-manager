package api

import (
	"net/http"
	"openshield-manager/internal/db"
	grpcclient "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"
	"openshield-manager/proto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssignTaskToAgentRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
	JobID   string `json:"job_id" binding:"required"`
}

func AssignTaskToAgent(c *gin.Context) {
	var req AssignTaskToAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", req.AgentID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Check if the job exists
	var job models.Job
	if err := db.DB.Where("id = ?", req.JobID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Store task in DB
	task := models.Task{
		ID:      uuid.New(),
		JobID:   job.ID,
		AgentID: agent.ID,
	}
	if err := db.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task: " + err.Error()})
		return
	}

	// Send task to agent
	client, err := grpcclient.NewAgentClient(agent.Address)
	if err != nil {
		// Handle the error appropriately, e.g., log or return a response
		// For now, just log and return
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gRPC client: " + err.Error()})
		return
	}
	// Use the client to send the task to the agent
	protoTask := &proto.Task{
		Id:      task.ID.String(),
		JobId:   task.JobID.String(),
		AgentId: task.AgentID.String(),
	}
	// Convert models.Job to proto.Job
	protoJob := &proto.Job{
		Id:          job.ID.String(),
		Name:        job.Name,
		Description: job.Description,
		Type:        string(job.Type),
		Target:      job.Target,
	}
	client.SendTask(c, protoTask, protoJob)

	c.JSON(http.StatusCreated, job)

	go grpcclient.TrackTaskStatus(agent.Address, task.ID.String(), job.ID.String(), 1*time.Second)

}

// GetTasksByAgent returns all tasks for a given agent ID
func GetTasksByAgent(c *gin.Context) {
	agentID := c.Param("id")

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", agentID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Fetch tasks for the agent
	var tasks []models.Task
	if err := db.DB.Where("agent_id = ?", agentID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks for agent: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetAllTasks returns all tasks in the system
func GetAllTasks(c *gin.Context) {
	var tasks []models.Task
	if err := db.DB.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}
