package api

import "github.com/gin-gonic/gin"

func CreateRouter() *gin.Engine {
	// Initialize the router
	router := gin.Default()

	// Serve static web interface files
	router.Static("/static", "./web")
	router.StaticFile("/", "./web/index.html")

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
	}

	return router
}
