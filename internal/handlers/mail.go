package handlers

import (
	"doodocs/internal/errors"
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
		errors.HandleErrorXML(w, "Method not allowed", http.StatusMethodNotAllowed, "Only POST method is allowed for sending emails.")
		return
	}

	if err := r.ParseMultipartForm(100000); err != nil {
		errors.HandleErrorXML(w, "Failed to parse form", http.StatusBadRequest, err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		errors.HandleErrorXML(w, "Failed to read file", http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	allowedMimeTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/pdf": true,
	}

	mimeType := header.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		errors.HandleErrorXML(w, "Invalid file type", http.StatusBadRequest, fmt.Sprintf("File type %s is not allowed", mimeType))
		return
	}

	tempFilePath, err := h.mailService.SaveFile(file, header)
	if err != nil {
		errors.HandleErrorXML(w, "Failed to save file", http.StatusInternalServerError, err.Error())
		return
	}
	defer h.mailService.RemoveFile(tempFilePath)

	emails := r.FormValue("emails")
	if emails == "" {
		errors.HandleErrorXML(w, "Emails are required", http.StatusBadRequest, "Please provide at least one email address.")
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
		errors.HandleErrorXML(w, "Failed to send email", http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}
