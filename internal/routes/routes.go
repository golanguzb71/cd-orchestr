package routes

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	connection := r.Group("/connection")
	connection.POST("/create")
	connection.POST("/close")
	fileFolder := r.Group("/file-foleder")
	fileFolder.GET("/path")
	fileFolder.POST("/create")
	fileFolder.PUT("/edit")
	fileFolder.DELETE("/delete")
}
