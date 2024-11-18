package handlers

import (
	"doodocs/internal/models"
	"doodocs/internal/services"
	"fmt"
	"net/http"
	"strings"
)

type MailHandler struct {
	mailService *services.MailService
}

func NewMailHandler(mailService *services.MailService) *MailHandler {
	return &MailHandler{mailService: mailService}
}

func (h *MailHandler) SendMail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(100000); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	allowedMimeTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/pdf": true,
	}

	mimeType := header.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		http.Error(w, fmt.Sprintf("File type %s is not allowed", mimeType), http.StatusBadRequest)
		return
	}

	tempFilePath, err := h.mailService.SaveFile(file, header)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.mailService.RemoveFile(tempFilePath)

	emails := r.FormValue("emails")
	if emails == "" {
		http.Error(w, "Emails are required", http.StatusBadRequest)
		return
	}
	emailList := strings.Split(emails, ",")

	mailDetails := &models.MailDetails{
		To:             emailList,
		Subject:        "Your requested file",
		Body:           "Please find the attached file.",
		AttachmentPath: tempFilePath,
	}

	err = h.mailService.SendMail(mailDetails)
	if err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}
