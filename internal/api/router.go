package api

import "github.com/gin-gonic/gin"

func CreateRouter() *gin.Engine {
	// Initialize the router
	router := gin.Default()
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
	}

	return router
}
