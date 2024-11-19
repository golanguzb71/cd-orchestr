package main

import (
	"github.com/gin-contrib/cors"
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

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},                                                       // Allow all origins, you can restrict this to specific domains
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                 // Allowed methods
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept", "x-session-id"}, // Allow x-session-id header
		AllowCredentials: true,
	}

	// Apply CORS middleware
	r.Use(cors.New(corsConfig))

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
