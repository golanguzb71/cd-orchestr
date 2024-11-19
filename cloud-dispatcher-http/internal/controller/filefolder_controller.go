package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	models "session-bridge/internal/model"
)

func GetPath(c *gin.Context) {
	fmt.Println(c.Get("connection"))
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	path := c.Query("path")
	files, err := fileFolderService.ListPath(conn, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

func CreateFolder(c *gin.Context) {
	itemType := c.Param("type")
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := fileFolderService.CreateFolder(conn, &req, itemType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item created successfully"})
}

func EditPath(c *gin.Context) {
	fileNewName := c.Param("name")
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.EditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := fileFolderService.EditPath(conn, &req, fileNewName)
	fmt.Println(err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Path edited successfully"})
}

func DeletePath(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	err := fileFolderService.DeletePath(conn, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Path deleted successfully"})
}

func OpenFile(ctx *gin.Context) {

}
