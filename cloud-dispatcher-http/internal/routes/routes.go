package routes

import (
	"github.com/gin-gonic/gin"
	"session-bridge/internal/controller"
	"session-bridge/internal/db"
	middleware "session-bridge/internal/utils"
)

func SetupRoutes(r *gin.Engine, redis *db.Redis) {
	connection := r.Group("/connection")
	{
		connection.POST("/create", controller.CreateConnection)
		connection.POST("/close", controller.CloseConnection)
	}
	fileFolder := r.Group("/file-folder")
	{
		fileFolder.Use(middleware.SessionAuth(redis))
		fileFolder.GET("/path", controller.GetPath)
		fileFolder.POST("/create/:type", controller.CreateFolder)
		fileFolder.PUT("/edit/:name", controller.EditPath)
		fileFolder.DELETE("/delete", controller.DeletePath)
		fileFolder.POST("/open-file", controller.OpenFile)
	}
	git := r.Group("/git")
	{
		git.Use(middleware.SessionAuth(redis))
		git.POST("/add", controller.GitAdd)
		git.POST("/commit", controller.GitCommit)
		git.POST("/switch-branch", controller.GitSwitchBranch)
		git.POST("/push", controller.GitPush)
		git.POST("/clone", controller.GitClone)
		git.POST("/pull", controller.GitPull)
	}
}
