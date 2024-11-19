package errors

import (
	"doodocs/internal/models"
	"encoding/xml"
	"net/http"
)

func HandleErrorXML(w http.ResponseWriter, errMsg string, statusCode int, description string) {
	errorResponse := models.ErrorResponse{
		Code:        statusCode,
		Message:     errMsg,
		Description: description,
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)
	xml.NewEncoder(w).Encode(errorResponse)
}

func IsValidMimeType(mimeType string) bool {
	allowedMimeTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/pdf": true,
		"application/xml": true,
		"image/jpeg":      true,
		"image/png":       true,
	}
	return allowedMimeTypes[mimeType]
}
