package repo

import "database/sql"

type ConnectionRepository struct {
	connectionDB *sql.DB
}

func NewConnectionRepository(db *sql.DB) *ConnectionRepository {
	return &ConnectionRepository{connectionDB: db}
}
