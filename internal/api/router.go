package api

import (
	"time"

	"openshield-manager/internal/events"
	"openshield-manager/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CreateRouter() *gin.Engine {
	// Initialize the router
	router := gin.Default()

	// Initialize event hub
	events.Init()

	// Configure CORS for dashboard
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://localhost:3000", "https://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// SSE endpoints for real-time events (org-aware, authenticated)
	router.GET("/events/agents", middleware.AuthMiddleware(), middleware.RequireOrgAccess(), events.HandleSSE)
	router.GET("/events/tasks", middleware.AuthMiddleware(), middleware.RequireOrgAccess(), events.HandleSSE)
	router.GET("/events/queries", middleware.AuthMiddleware(), middleware.RequireOrgAccess(), events.HandleSSE)

	// External API routes
	apiGroup := router.Group("/api")
	{
		agents := apiGroup.Group("/agents")
		agents.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			agents.POST("/unregister", middleware.RequireRole(middleware.RoleAdmin), UnregisterAgent)
			agents.GET("/list", middleware.RequireRole(middleware.RoleViewer), GetAgentsList)
			agents.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetAgentDetails)
			agents.GET("/:id/tasks", middleware.RequireRole(middleware.RoleViewer), GetTasksByAgent)
			agents.GET("/:id/tools", middleware.RequireRole(middleware.RoleViewer), GetToolsByAgent)
		}
		// Jobs endpoints
		jobs := apiGroup.Group("/jobs")
		jobs.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			jobs.GET("/list", middleware.RequireRole(middleware.RoleViewer), GetJobs)
			jobs.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetJobDetails)
			jobs.POST("/create", middleware.RequireRole(middleware.RoleOperator), CreateJob)
		}
		// Tasks endpoints
		tasks := apiGroup.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			tasks.POST("/assign", middleware.RequireRole(middleware.RoleOperator), AssignTaskToAgent)
			tasks.GET("/list", middleware.RequireRole(middleware.RoleViewer), GetAllTasks)
		}
		// Tools endpoints
		tools := apiGroup.Group("/tools")
		tools.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			tools.POST("/execute", middleware.RequireRole(middleware.RoleOperator), ExecuteTool)
		}
		// Certificates endpoints
		// Note: cert/sign uses agent token (X-Agent-Token) auth, not JWT,
		// because agents don't have dashboard credentials during enrollment.
		cert := apiGroup.Group("/certs")
		{
			cert.POST("/sign", SignAgentCSR)
		}
		// Also register at /cert/sign (singular) for backward compatibility
		apiGroup.POST("/cert/sign", SignAgentCSR)
		// Queries endpoints - FleetDM-style
		queries := apiGroup.Group("/queries")
		queries.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			queries.GET("/list", middleware.RequireRole(middleware.RoleViewer), GetQueries)
			queries.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetQuery)
			queries.POST("/create", middleware.RequireRole(middleware.RoleOperator), CreateQuery)
			queries.PUT("/:id", middleware.RequireRole(middleware.RoleOperator), UpdateQuery)
			queries.DELETE("/:id", middleware.RequireRole(middleware.RoleAdmin), DeleteQuery)
			queries.POST("/run", middleware.RequireRole(middleware.RoleOperator), RunQuery)
			queries.POST("/run-live", middleware.RequireRole(middleware.RoleOperator), RunLiveQuery)
		}
		// Query executions endpoints
		executions := apiGroup.Group("/query-executions")
		executions.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			executions.GET("/list", middleware.RequireRole(middleware.RoleViewer), GetQueryExecutions)
			executions.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetQueryExecution)
		}

		// Agent Groups endpoints
		groups := apiGroup.Group("/groups")
		groups.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			groups.GET("", middleware.RequireRole(middleware.RoleViewer), GetGroups)
			groups.POST("", middleware.RequireRole(middleware.RoleOperator), CreateGroup)
			groups.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetGroup)
			groups.PUT("/:id", middleware.RequireRole(middleware.RoleOperator), UpdateGroup)
			groups.DELETE("/:id", middleware.RequireRole(middleware.RoleAdmin), DeleteGroup)
			groups.POST("/:id/agents", middleware.RequireRole(middleware.RoleOperator), AddAgentsToGroup)
			groups.DELETE("/:id/agents", middleware.RequireRole(middleware.RoleOperator), RemoveAgentsFromGroup)
		}

		// Bulk Operations endpoints
		bulkOps := apiGroup.Group("/bulk-operations")
		bulkOps.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
		{
			bulkOps.GET("", middleware.RequireRole(middleware.RoleViewer), GetBulkOperations)
			bulkOps.POST("", middleware.RequireRole(middleware.RoleOperator), CreateBulkOperation)
			bulkOps.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetBulkOperation)
			bulkOps.POST("/:id/cancel", middleware.RequireRole(middleware.RoleAdmin), CancelBulkOperation)
		}

		// Auth endpoints (Basic Auth → JWT, Wazuh-style)
		RegisterAuthRoutes(apiGroup)

		// Organization management endpoints
		RegisterOrgRoutes(apiGroup)

		// Registration token management endpoints
		RegisterRegTokenRoutes(apiGroup)
	}

	return router
}
