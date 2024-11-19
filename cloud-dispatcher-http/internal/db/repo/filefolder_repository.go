package repo

import "database/sql"

type FileFolderRepository struct {
	db *sql.DB
}

func NewFileFolderRepository(db *sql.DB) *FileFolderRepository {
	return &FileFolderRepository{db: db}
}
