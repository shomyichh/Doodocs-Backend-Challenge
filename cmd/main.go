package main

import (
	"doodocs/internal/config"
	"doodocs/internal/handlers"
	"doodocs/internal/repositories"
	"doodocs/internal/services"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {

	config.Init()

	tempDir := config.Get("TMPDIR")
	smtpHost := config.Get("SMTP_HOST")
	smtpUser := config.Get("SMTP_USER")
	smtpPass := config.Get("SMTP_PASS")
	smtpPort, err := strconv.Atoi(config.Get("SMTP_PORT"))

	if err != nil {
		fmt.Println("Error with readin port")
	}

	fileRepo := repositories.NewFileRepository(tempDir)
	archiveRepo := repositories.NewArchiveRepository()
	mailRepo := repositories.NewMailRepository(smtpHost, smtpPort, smtpUser, smtpPass)

	archiveService := services.NewArchiveService(fileRepo, archiveRepo)
	mailService := services.NewMailService(mailRepo)

	archiveHandler := handlers.NewArchiveHandler(archiveService)
	mailHandler := handlers.NewMailHandler(mailService)

	http.HandleFunc("/api/archive/information", archiveHandler.GetArchiveInformation)
	http.HandleFunc("/api/archive/files", archiveHandler.CreateArchive)
	http.HandleFunc("/api/mail/file", mailHandler.SendMail)

	port := ":8080"
	log.Printf("Server is running on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
