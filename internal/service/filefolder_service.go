package service

import "session-bridge/internal/db/repo"

type FileFolderService struct {
	fileFolderRepo *repo.FileFolderRepository
}

func NewFileFolderService(fileFolderRepo *repo.FileFolderRepository) *FileFolderService {
	return &FileFolderService{fileFolderRepo: fileFolderRepo}
}
