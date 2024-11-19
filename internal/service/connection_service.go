package service

import (
	"context"
	"golang.org/x/crypto/ssh"
	"session-bridge/internal/db/repo"
	models "session-bridge/internal/model"
)

type ConnectionService struct {
	connectionRepo *repo.ConnectionRepository
}

func NewConnectionService(connectionRepo *repo.ConnectionRepository) *ConnectionService {
	return &ConnectionService{
		connectionRepo: connectionRepo,
	}
}

func (s *ConnectionService) CreateConnection(ctx context.Context, req *models.ConnectionRequest) (*models.ConnectionResponse, error) {
	conn := &models.ServerConnection{
		IPAddress: req.IPAddress,
		Username:  req.Username,
		Password:  req.Password,
		Port:      req.Port,
	}

	sessionID, client, err := s.connectionRepo.CreateConnection(ctx, conn)
	if err != nil {
		return nil, err
	}
	go s.monitorConnection(sessionID, client)

	return &models.ConnectionResponse{
		SessionID: sessionID,
		Message:   "SSH Connection established successfully",
	}, nil
}

func (s *ConnectionService) CloseConnection(ctx context.Context, sessionID string) error {
	return s.connectionRepo.CloseConnection(ctx, sessionID)
}

func (s *ConnectionService) monitorConnection(sessionID string, client *ssh.Client) {
	go func() {
		client.Wait()
		ctx := context.Background()
		_ = s.connectionRepo.CloseConnection(ctx, sessionID)
	}()
}
