package repositories

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

type MailRepository struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

func NewMailRepository(host string, port int, user, password string) *MailRepository {
	return &MailRepository{
		SMTPHost:     host,
		SMTPPort:     port,
		SMTPUser:     user,
		SMTPPassword: password,
	}
}

type Email struct {
	To          []string
	Subject     string
	Body        string
	Attachments map[string][]byte
}

func (repo *MailRepository) SendEmail(email *Email) error {
	if len(email.To) == 0 {
		return errors.New("recipient list cannot be empty")
	}

	from := repo.SMTPUser
	password := repo.SMTPPassword

	message, err := repo.buildMessage(email)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", repo.SMTPHost, repo.SMTPPort)
	auth := smtp.PlainAuth("", from, password, repo.SMTPHost)

	return smtp.SendMail(addr, auth, from, email.To, message)
}

func (repo *MailRepository) buildMessage(email *Email) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	buf.WriteString(fmt.Sprintf("From: %s\r\n", repo.SMTPUser))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ",")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary()))

	textPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain; charset=utf-8"},
	})
	if err != nil {
		return nil, err
	}
	textPart.Write([]byte(email.Body))

	for filename, content := range email.Attachments {
		attachmentPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {http.DetectContentType(content)},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", filename)},
		})
		if err != nil {
			return nil, err
		}

		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(content)))
		base64.StdEncoding.Encode(encoded, content)
		attachmentPart.Write(encoded)
	}

	// Закрываем MIME writer
	writer.Close()
	return buf.Bytes(), nil
}

// SaveFile сохраняет файл временно
func (repo *MailRepository) SaveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	filePath := "./tmp/" + header.Filename

	// Создаем временную директорию, если её нет
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		err = os.Mkdir("./tmp", os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = file.Seek(0, 0) // Сброс указателя чтения файла
	if err != nil {
		return "", err
	}

	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// AttachFile загружает содержимое файла для использования в качестве вложения
func (repo *MailRepository) AttachFile(filePath string) (map[string][]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	_, filename := filepath.Split(filePath)
	return map[string][]byte{filename: content}, nil
}
