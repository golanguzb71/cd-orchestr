package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"session-bridge/internal/db"
	models "session-bridge/internal/model"
)

const sessionKeyFormat = "session:%s"

func SessionAuth(redis *db.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session ID is required"})
			c.Abort()
			return
		}

		sessionKey := fmt.Sprintf(sessionKeyFormat, sessionID)
		sessionData, err := redis.Get(context.TODO(), sessionKey)
		if err != nil || sessionData == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		var sessionInfo models.SessionInfo
		if err := json.Unmarshal([]byte(sessionData), &sessionInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse session data"})
			c.Abort()
			return
		}

		c.Set("connection", sessionInfo.Connection)
		c.Next()
	}
}
