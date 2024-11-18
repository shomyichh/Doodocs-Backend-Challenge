package handlers

import (
	"doodocs/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ArchiveHandler struct {
	archiveService *services.ArchiveService
}

func NewArchiveHandler(archiveService *services.ArchiveService) *ArchiveHandler {
	return &ArchiveHandler{archiveService: archiveService}
}

func (h *ArchiveHandler) GetArchiveInformation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}
	if len(files) > 1 {
		http.Error(w, "Only one file should be uploaded", http.StatusBadRequest)
		return
	}

	fileHeader := files[0]
	if !strings.HasSuffix(fileHeader.Filename, ".zip") {
		http.Error(w, "The uploaded file is not a .ZIP file", http.StatusBadRequest)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		http.Error(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	archiveInfo, err := h.archiveService.ProcessArchive(file, fileHeader)
	if err != nil {
		http.Error(w, "Failed to process archive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(archiveInfo); err != nil {
		http.Error(w, "Failed to encode archive info: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ArchiveHandler) CreateArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(1000000); err != nil {
		return
	}

	files := r.MultipartForm.File["files[]"]

	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	allowedMimeTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/xml": true,
		"image/jpeg":      true,
		"image/png":       true,
	}

	var filePaths []string
	var invalidFiles []string
	for _, fileHeader := range files {

		mimeType := fileHeader.Header.Get("Content-Type")
		if !allowedMimeTypes[mimeType] {
			invalidFiles = append(invalidFiles, fileHeader.Filename)

		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		filePath, err := h.archiveService.SaveFile(file, fileHeader)
		if err != nil {
			http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		filePaths = append(filePaths, filePath)

	}

	if len(invalidFiles) > 0 {
		errorMessage := fmt.Sprintf("Invalid file formats for: %v", invalidFiles)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	outputFilePath, err := h.archiveService.CreateArchive(filePaths)
	if err != nil {
		http.Error(w, "Failed to create archive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")

	http.ServeFile(w, r, outputFilePath)

	// Удаление временного архива после отправки
	defer os.Remove(outputFilePath)
}
