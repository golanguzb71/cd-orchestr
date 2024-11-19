package service

import "session-bridge/internal/db/repo"

type ConnectionService struct {
	connectionRepo *repo.ConnectionRepository
}

func NewConnectionService(connectionRepo *repo.ConnectionRepository) *ConnectionService {
	return &ConnectionService{connectionRepo: connectionRepo}
}
