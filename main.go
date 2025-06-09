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

	// Generate manager certificates
	err = service.CreateCertificates()
	if err != nil {
		log.Fatalf("Failed to create certificates: %v", err)
	}

	// Start the RegisterAgent gRPC server
	go func() {
		err := managergrpc.StartManagerRegistrationServer(50053)
		if err != nil {
			log.Fatalf("Failed to start Manager Registration Server server: %v", err)
		}
	}()

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
	stopConfigsSync := make(chan struct{})
	service.ConfigSyncMonitor(60*time.Second, stopConfigsSync)
	stopAgentMonitor := make(chan struct{})
	service.AgentLastSeenMonitor(30*time.Second, stopAgentMonitor)

	// Start the API router
	router := api.CreateRouter()
	router.Run(":9000")
}
