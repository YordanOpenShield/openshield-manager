package main

import (
	"openshield-manager/internal/api"
	"openshield-manager/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.ConnectDatabase()
	router := gin.Default()

	apiGroup := router.Group("/api")
	{
		agent := apiGroup.Group("/agents")
		{
			agent.POST("/register", api.RegisterAgent)
			agent.POST("/unregister", api.UnregisterAgent)
			agent.POST("/heartbeat", api.AgentHeartbeat)
		}
		jobs := apiGroup.Group("/jobs")
		{
			jobs.GET("/available", api.GetAvailableJobs)
		}
	}

	router.Run(":9000")
}
