package main

import (
	"log"
	"openshield-manager/internal/api"
	"openshield-manager/internal/config"
	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

const configFile = "config/config.yml"

func main() {
	// Load the configuration file
	err := config.LoadAndSetConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize the database connection
	db.ConnectDatabase()

	// Start the gRPC server in a goroutine
	go func() {
		err := managergrpc.StartGRPCServer(50052)
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start background tasks
	stopScriptsSync := make(chan struct{})
	service.ScriptSyncMonitor(60*time.Second, stopScriptsSync)
	stopAgentMonitor := make(chan struct{})
	service.AgentLastSeenMonitor(30*time.Second, stopAgentMonitor)

	// Initialize the router
	router := gin.Default()
	// External API routes
	apiGroup := router.Group("/api")
	{
		agents := apiGroup.Group("/agents")
		{
			agents.POST("/unregister", api.UnregisterAgent)
			agents.GET("/list", api.GetAgentsList)
			agents.GET("/:id", api.GetAgentDetails)
			agents.GET("/:id/tasks", api.GetTasksByAgent)
		}
		// Jobs endpoints
		jobs := apiGroup.Group("/jobs")
		{
			jobs.GET("/list", api.GetJobs)
			jobs.GET("/:id", api.GetJobDetails)
			jobs.POST("/create", api.CreateJob)
		}
		// Tasks endpoints
		tasks := apiGroup.Group("/tasks")
		{
			tasks.POST("/assign", api.AssignTaskToAgent)
			tasks.GET("/list", api.GetAllTasks)
		}
	}

	router.Run(":9000")
}
