package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"openshield-manager/internal/events"
)

func CreateRouter() *gin.Engine {
	// Initialize the router
	router := gin.Default()

	// Initialize event hub
	events.Init()

	// Configure CORS for dashboard
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// SSE endpoints for real-time events
	router.GET("/events/agents", events.HandleSSE)
	router.GET("/events/tasks", events.HandleSSE)
	router.GET("/events/queries", events.HandleSSE)

	// External API routes
	apiGroup := router.Group("/api")
	{
		agents := apiGroup.Group("/agents")
		{
			agents.POST("/unregister", UnregisterAgent)
			agents.GET("/list", GetAgentsList)
			agents.GET("/:id", GetAgentDetails)
			agents.GET("/:id/tasks", GetTasksByAgent)
			agents.GET("/:id/tools", GetToolsByAgent)
		}
		// Jobs endpoints
		jobs := apiGroup.Group("/jobs")
		{
			jobs.GET("/list", GetJobs)
			jobs.GET("/:id", GetJobDetails)
			jobs.POST("/create", CreateJob)
		}
		// Tasks endpoints
		tasks := apiGroup.Group("/tasks")
		{
			tasks.POST("/assign", AssignTaskToAgent)
			tasks.GET("/list", GetAllTasks)
		}
		// Tools endpoints
		tools := apiGroup.Group("/tools")
		{
			tools.POST("/execute", ExecuteTool)
		}
		// Certificates endpoints
		cert := apiGroup.Group("/certs")
		{
			cert.POST("/sign", SignAgentCSR)
		}
		// Queries endpoints - FleetDM-style
		queries := apiGroup.Group("/queries")
		{
			queries.GET("/list", GetQueries)
			queries.GET("/:id", GetQuery)
			queries.POST("/create", CreateQuery)
			queries.PUT("/:id", UpdateQuery)
			queries.DELETE("/:id", DeleteQuery)
			queries.POST("/run", RunQuery)
			queries.POST("/run-live", RunLiveQuery)
		}
		// Query executions endpoints
		executions := apiGroup.Group("/query-executions")
		{
			executions.GET("/list", GetQueryExecutions)
			executions.GET("/:id", GetQueryExecution)
		}

		// Agent Groups endpoints
		groups := apiGroup.Group("/groups")
		{
			groups.GET("", GetGroups)
			groups.POST("", CreateGroup)
			groups.GET("/:id", GetGroup)
			groups.PUT("/:id", UpdateGroup)
			groups.DELETE("/:id", DeleteGroup)
			groups.POST("/:id/agents", AddAgentsToGroup)
			groups.DELETE("/:id/agents", RemoveAgentsFromGroup)
		}

		// Bulk Operations endpoints
		bulkOps := apiGroup.Group("/bulk-operations")
		{
			bulkOps.GET("", GetBulkOperations)
			bulkOps.POST("", CreateBulkOperation)
			bulkOps.GET("/:id", GetBulkOperation)
			bulkOps.POST("/:id/cancel", CancelBulkOperation)
		}
	}

	return router
}
