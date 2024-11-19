package handlers

import (
	"doodocs/internal/errors"
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
		errors.HandleErrorXML(w, "Method not allowed", http.StatusMethodNotAllowed, "Only POST method is allowed for this operation.")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		errors.HandleErrorXML(w, "Failed to parse multipart form", http.StatusBadRequest, err.Error())
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		errors.HandleErrorXML(w, "No files uploaded", http.StatusBadRequest, "You must upload a .zip file.")
		return
	}
	if len(files) > 1 {
		errors.HandleErrorXML(w, "Too many files", http.StatusBadRequest, "Only one file should be uploaded.")
		return
	}

	fileHeader := files[0]
	if !strings.HasSuffix(fileHeader.Filename, ".zip") {
		errors.HandleErrorXML(w, "Invalid file type", http.StatusBadRequest, "The uploaded file is not a .ZIP file.")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		errors.HandleErrorXML(w, "Failed to open file", http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	archiveInfo, err := h.archiveService.ProcessArchive(file, fileHeader)
	if err != nil {
		errors.HandleErrorXML(w, "Failed to process archive", http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(archiveInfo); err != nil {
		errors.HandleErrorXML(w, "Failed to encode archive info", http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *ArchiveHandler) CreateArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.HandleErrorXML(w, "Method not allowed", http.StatusMethodNotAllowed, "Only POST method is allowed for this operation.")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		errors.HandleErrorXML(w, "Failed to parse multipart form", http.StatusBadRequest, err.Error())
		return
	}

	files := r.MultipartForm.File["files[]"]
	if len(files) == 0 {
		errors.HandleErrorXML(w, "No files uploaded", http.StatusBadRequest, "At least one file must be uploaded.")
		return
	}

	var filePaths []string
	var invalidFiles []string
	for _, fileHeader := range files {
		mimeType := fileHeader.Header.Get("Content-Type")
		if !errors.IsValidMimeType(mimeType) {
			invalidFiles = append(invalidFiles, fileHeader.Filename)
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			errors.HandleErrorXML(w, "Failed to open file", http.StatusBadRequest, err.Error())
			return
		}
		defer file.Close()

		filePath, err := h.archiveService.SaveFile(file, fileHeader)
		if err != nil {
			errors.HandleErrorXML(w, "Failed to save file", http.StatusInternalServerError, err.Error())
			return
		}
		filePaths = append(filePaths, filePath)
	}

	if len(invalidFiles) > 0 {
		errorMessage := fmt.Sprintf("Invalid file formats for: %v", invalidFiles)
		errors.HandleErrorXML(w, errorMessage, http.StatusBadRequest, "Some files have unsupported formats.")
		return
	}

	outputFilePath, err := h.archiveService.CreateArchive(filePaths)
	if err != nil {
		errors.HandleErrorXML(w, "Failed to create archive", http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")
	http.ServeFile(w, r, outputFilePath)

	// Удаление временного архива после отправки
	defer os.RemoveAll(outputFilePath)
}
