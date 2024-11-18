package models

type MailDetails struct {
	To             []string `json:"to"`
	Subject        string   `json:"subject"`
	Body           string   `json:"body"`
	AttachmentPath string   `json:"attachment_path"`
}
