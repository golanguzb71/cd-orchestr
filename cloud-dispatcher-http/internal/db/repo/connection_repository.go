package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"session-bridge/internal/db"
	models "session-bridge/internal/model"
	"time"
)

type ConnectionRepository struct {
	connectionDB *sql.DB
	redisDB      *db.Redis
}

func NewConnectionRepository(db *sql.DB, redisClient *db.Redis) *ConnectionRepository {
	return &ConnectionRepository{
		connectionDB: db,
		redisDB:      redisClient,
	}
}

func (r *ConnectionRepository) CreateConnection(ctx context.Context, conn *models.ServerConnection) (string, *ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: conn.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(conn.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	port := 22
	if conn.Port != 0 {
		port = conn.Port
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.IPAddress, port), config)
	if err != nil {
		return "", nil, fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	query := `INSERT INTO server_connections (ip_address, username, password, port) VALUES ($1, $2, $3, $4)`
	result, err := r.connectionDB.ExecContext(ctx, query, conn.IPAddress, conn.Username, conn.Password, port)
	if err != nil {
		client.Close()
		return "", nil, fmt.Errorf("failed to insert connection: %v", err)
	}

	conn.ID, _ = result.LastInsertId()

	sessionID := uuid.New().String()

	sessionInfo := &models.SessionInfo{
		Connection: conn,
		CreatedAt:  time.Now(),
	}

	sessionJSON, err := json.Marshal(sessionInfo)
	if err != nil {
		client.Close()
		return "", nil, fmt.Errorf("failed to marshal session info: %v", err)
	}

	err = r.redisDB.Set(ctx, fmt.Sprintf("session:%s", sessionID), sessionJSON, 1*time.Hour)
	if err != nil {
		client.Close()
		return "", nil, fmt.Errorf("failed to store in Redis: %v", err)
	}

	return sessionID, client, nil
}

func (r *ConnectionRepository) CloseConnection(ctx context.Context, sessionID string) error {
	_, err := r.redisDB.Get(ctx, fmt.Sprintf("session:%s", sessionID))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session from Redis: %v", err)
	}

	err = r.redisDB.Del(ctx, fmt.Sprintf("session:%s", sessionID))
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %v", err)
	}

	return nil
}
