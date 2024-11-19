package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	models "session-bridge/internal/model"
	"session-bridge/internal/service"
)

var (
	connService       *service.ConnectionService
	fileFolderService *service.FileFolderService
)

func SetConnectionService(connSer *service.ConnectionService, ffService *service.FileFolderService) {
	connService = connSer
	fileFolderService = ffService
}

func CreateConnection(c *gin.Context) {
	var req models.ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := connService.CreateConnection(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func CloseConnection(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	err := connService.CloseConnection(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SSH Connection closed successfully"})
}
