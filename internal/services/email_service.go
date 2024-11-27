package services

import (
	"fmt"
	"net/smtp"
)

type EmailService struct {
	smtpHost string
	smtpPort string
	username string
	password string
}

func NewEmailService(smtpHost, smtpPort, username, password string) *EmailService {
	return &EmailService{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
	}
}

func (s *EmailService) SendVerificationEmail(to, token string) error {
	from := s.username
	auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)

	link := fmt.Sprintf("http://localhost:8080/verify-email?token=%s", token)
	body := fmt.Sprintf("Click the link to verify your email: %s", link)
	message := []byte("Subject: Email Verification\r\n\r\n" + body)

	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
