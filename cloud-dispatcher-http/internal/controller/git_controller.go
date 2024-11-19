package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	models "session-bridge/internal/model"
)

func GitClone(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.GitCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := gitService.CloneRepo(conn, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone repository", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git clone successful"})
}

func GitPull(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.GitPullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := gitService.PullRepo(conn, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pull repository", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git pull successful"})
}

func GitPush(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.GitPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := gitService.PushRepo(conn, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push changes", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git push successful"})
}

func GitAdd(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.GitAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := gitService.AddFiles(conn, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute git add", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git add successful"})
}

func GitCommit(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.GitCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := gitService.CommitChanges(conn, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute git commit", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git commit successful"})
}

func GitSwitchBranch(c *gin.Context) {
	conn, ok := c.MustGet("connection").(*models.ServerConnection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve connection from context"})
		return
	}

	var req models.GitSwitchBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := gitService.SwitchBranch(conn, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to switch git branch", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Switched git branch successfully"})
}
