package main

import (
	"openshield-manager/internal/api"
	"openshield-manager/internal/db"
	"openshield-manager/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database connection
	db.ConnectDatabase()

	// Start background tasks
	stopCh := make(chan struct{})
	service.StartGlobalScriptSyncMonitor(30*time.Second, stopCh)

	// Initialize the router
	router := gin.Default()
	apiGroup := router.Group("/api")
	{
		// Register endpoint (no auth)
		apiGroup.POST("/agents/register", api.RegisterAgent)
		// Authenticated agents endpoints
		agent := apiGroup.Group("/agents", api.AgentAuthMiddleware())
		{
			agent.POST("/unregister", api.UnregisterAgent)
			agent.POST("/heartbeat", api.AgentHeartbeat)
		}
		// Jobs endpoints
		jobs := apiGroup.Group("/jobs")
		{
			jobs.GET("/available", api.GetAvailableJobs)
			jobs.POST("/create", api.CreateJob)
		}
		// Tasks endpoints
		tasks := apiGroup.Group("/tasks")
		{
			tasks.POST("/assign", api.AssignTaskToAgent)
		}
	}

	router.Run(":9000")
}
