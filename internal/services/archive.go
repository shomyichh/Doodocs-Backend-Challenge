package services

import (
	"doodocs/internal/models"
	"doodocs/internal/repositories"
	"mime/multipart"
)

type ArchiveService struct {
	fileRepo    *repositories.FileRepository
	archiveRepo *repositories.ArchiveRepository
}

func NewArchiveService(fileRepo *repositories.FileRepository, archiveRepo *repositories.ArchiveRepository) *ArchiveService {
	return &ArchiveService{
		fileRepo:    fileRepo,
		archiveRepo: archiveRepo,
	}
}
func (s *ArchiveService) SaveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	return s.fileRepo.SaveFile(file, header)
}

func (s *ArchiveService) ProcessArchive(file multipart.File, header *multipart.FileHeader) (*models.ArchiveInfo, error) {
	tempFilePath, err := s.fileRepo.SaveFile(file, header)
	if err != nil {
		return nil, err
	}
	defer s.fileRepo.RemoveFile(tempFilePath)

	files, totalSize, err := s.archiveRepo.ExtractArchiveInfo(tempFilePath)
	if err != nil {
		return nil, err
	}

	archiveSize, err := s.fileRepo.GetFileSize(tempFilePath)
	if err != nil {
		return nil, err
	}

	return &models.ArchiveInfo{
		Filename:    header.Filename,
		ArchiveSize: float64(archiveSize),
		TotalSize:   totalSize,
		TotalFiles:  len(files),
		Files:       files,
	}, nil
}

func (s *ArchiveService) CreateArchive(files []string) (string, error) {
	_, err := s.fileRepo.CreateTempFolder()
	if err != nil {
		return "", err
	}

	outputFilePath := s.fileRepo.GetArchivePath()

	err = s.archiveRepo.CreateArchive(files, outputFilePath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		err := s.fileRepo.RemoveFile(file)
		if err != nil {
			//fmt.Println("Error removing file:", err)
		}
	}

	return outputFilePath, nil
}
