package main

import (
	"flag"
	"log"
	"net"
	"net/http"
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

	// Seed default organization and super admin if first run
	if err := service.SeedDefaultOrg(); err != nil {
		log.Printf("[MANAGER] Warning: seed failed (may already be seeded): %v", err)
	}

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
	httpAddr := ":" + config.GlobalConfig.HTTP_PORT

	if config.GlobalConfig.TLS_ENABLED {
		// HTTPS server on the configured HTTPS port
		httpsAddr := ":" + config.GlobalConfig.HTTPS_PORT
		tlsConfig, err := utils.LoadRESTTLSCredentials()
		if err != nil {
			log.Fatalf("Failed to load REST TLS credentials: %v", err)
		}
		httpsServer := &http.Server{
			Addr:      httpsAddr,
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		go func() {
			log.Printf("[MANAGER] REST API listening on %s (TLS enabled)", httpsAddr)
			if err := httpsServer.ListenAndServeTLS(config.CertsPath+"/manager.crt", config.CertsPath+"/manager.key"); err != nil {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		}()

		// HTTP server on port 9000: redirects all traffic to HTTPS
		httpServer := &http.Server{
			Addr: httpAddr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				host := r.Host
				// Replace the port if present
				if _, _, err := net.SplitHostPort(host); err == nil {
					host, _, _ = net.SplitHostPort(host)
				}
				redirectURL := "https://" + host + ":" + config.GlobalConfig.HTTPS_PORT + r.URL.RequestURI()
				http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			}),
		}
		log.Printf("[MANAGER] HTTP redirect on %s → HTTPS%s", httpAddr, httpsAddr)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start HTTP redirect server: %v", err)
		}
	} else {
		log.Printf("[MANAGER] REST API listening on %s (no TLS)", httpAddr)
		router.Run(httpAddr)
	}
}
