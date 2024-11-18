package services

import (
	"doodocs/internal/models"
	"doodocs/internal/repositories"
	"fmt"
	"mime/multipart"
	"os"
)

type MailService struct {
	mailRepo *repositories.MailRepository
}

func NewMailService(mailRepo *repositories.MailRepository) *MailService {
	return &MailService{mailRepo: mailRepo}
}

func (s *MailService) SaveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	return s.mailRepo.SaveFile(file, header)
}

func (s *MailService) RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

func (s *MailService) SendMail(details *models.MailDetails) error {
	email := &repositories.Email{
		To:      details.To,
		Subject: details.Subject,
		Body:    details.Body,
	}

	if details.AttachmentPath != "" {
		attachments, err := s.mailRepo.AttachFile(details.AttachmentPath)
		if err != nil {
			return fmt.Errorf("failed to attach file: %w", err)
		}
		email.Attachments = attachments
	}

	return s.mailRepo.SendEmail(email)
}
