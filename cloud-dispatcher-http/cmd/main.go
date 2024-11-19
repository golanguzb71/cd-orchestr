package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"session-bridge/internal/config"
	"session-bridge/internal/controller"
	"session-bridge/internal/db"
	"session-bridge/internal/db/repo"
	"session-bridge/internal/routes"
	"session-bridge/internal/service"
)

// @title Tender Managment Swagger

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	r := gin.Default()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error while loading config %v", err)
	}
	database := db.NewDatabase(&cfg.Database)
	redis := db.NewRedisClient(&cfg.Database)
	_ = repo.NewFileFolderRepository(database)
	ffService := service.NewFileFolderService(redis)
	connectionRepo := repo.NewConnectionRepository(database, redis)
	connectionService := service.NewConnectionService(connectionRepo)
	gitService := service.NewGitService(redis)
	controller.SetConnectionService(connectionService, ffService, gitService)
	routes.SetupRoutes(r, redis)
	r.Run(":8888")
}
