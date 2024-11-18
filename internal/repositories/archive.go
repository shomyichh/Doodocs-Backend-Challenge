package repositories

import (
	"archive/zip"
	"doodocs/internal/models"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
)

type ArchiveRepository struct{}

func NewArchiveRepository() *ArchiveRepository {
	return &ArchiveRepository{}
}

func (repo *ArchiveRepository) ExtractArchiveInfo(filePath string) ([]models.FileDetails, float64, error) {
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, 0, errors.New("not a valid ZIP archive")
	}
	defer zipReader.Close()

	var files []models.FileDetails
	var totalSize float64

	for _, f := range zipReader.File {
		totalSize += float64(f.UncompressedSize64)
		files = append(files, models.FileDetails{
			FilePath: f.Name,
			Size:     float64(f.UncompressedSize64),
			MimeType: mime.TypeByExtension(filepath.Ext(f.Name)),
		})
	}

	return files, totalSize, nil
}

func (repo *ArchiveRepository) CreateArchive(files []string, outputFilePath string) error {
	archiveFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	for _, file := range files {
		err := repo.addFileToArchive(zipWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *ArchiveRepository) addFileToArchive(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
