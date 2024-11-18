package repositories

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
)

type FileRepository struct {
	tempDir string
}

func NewFileRepository(tempDir string) *FileRepository {
	return &FileRepository{tempDir: tempDir}
}

func (repo *FileRepository) CreateTempFolder() (string, error) {
	if _, err := os.Stat(repo.tempDir); os.IsNotExist(err) {
		err := os.MkdirAll(repo.tempDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create temp folder: %w", err)
		}
	}
	return repo.tempDir, nil
}

func (repo *FileRepository) GetArchivePath() string {
	return path.Join(repo.tempDir, "output.zip")
}

func (repo *FileRepository) SaveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	_, err := repo.CreateTempFolder()
	if err != nil {
		return "", err
	}

	tempFile, err := os.CreateTemp(repo.tempDir, "uploaded-*-"+filepath.Base(header.Filename))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	return tempFile.Name(), nil
}

func (repo *FileRepository) RemoveFile(path string) error {
	return os.Remove(path)
}

func (repo *FileRepository) GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}
