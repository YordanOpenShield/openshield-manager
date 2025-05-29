package main

import (
	"flag"
	"log"
	"openshield-manager/internal/api"
	"openshield-manager/internal/config"
	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/service"
	"openshield-manager/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command-line arguments
	configPath := flag.String("config", config.ConfigPath, "Path to configuration file")
	scriptsPath := flag.String("scripts", config.ScriptsPath, "Path to scripts directory")
	certsPath := flag.String("certs", config.CertsPath, "Path to certificates directory")
	flag.Parse()
	config.ConfigPath = *configPath
	config.ScriptsPath = *scriptsPath
	config.CertsPath = *certsPath

	// Create the config directory if it doesn't exist
	utils.CreateConfig(config.ConfigPath, config.Config{})
	// Create the scripts directory if it doesn't exist
	utils.CreateScriptsDir(config.ScriptsPath)
	// Create the certs directory if it doesn't exist
	utils.CreateCertsDir(config.CertsPath)

	// Load the configuration file
	err := config.LoadAndSetConfig(config.ConfigPath)
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
		// Certificates endpoints
		cert := apiGroup.Group("/cert")
		{
			cert.POST("/sign", api.SignAgentCSR)
		}
	}

	router.Run(":9000")
}
