package models

import (
	"golang.org/x/crypto/ssh"
	"time"
)

type ServerConnection struct {
	ID        int64  `json:"id"`
	IPAddress string `json:"ip_address"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Port      int    `json:"port"`
}
type ConnectionRequest struct {
	IPAddress string `json:"ip_address" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Port      int    `json:"port" default:"22"`
}

type ConnectionResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type SessionInfo struct {
	Connection *ServerConnection `json:"connection"`
	Client     *ssh.Client       `json:"-"`
	CreatedAt  time.Time         `json:"created_at"`
}
