package services

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"
)

type EmailTask struct {
	To    string
	Token string
}

type EmailService struct {
	smtpHost string
	smtpPort string
	username string
	password string
	tasks    chan EmailTask // 작업 큐
}

func NewEmailService(smtpHost, smtpPort, username, password string, queueSize int, numWorkers int) *EmailService {
	service := &EmailService{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		tasks:    make(chan EmailTask, queueSize), // 큐 생성
	}

	// 워커 고루틴 실행
	for i := 0; i < numWorkers; i++ {
		go service.startWorker()
	}

	return service
}

// 워커 실행: 큐에서 작업을 처리
func (s *EmailService) startWorker() {
	for task := range s.tasks {
		err := s.sendEmail(task.To, task.Token)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", task.To, err)
		} else {
			log.Printf("Email sent successfully to %s", task.To)
		}
	}
}

var ErrEmailQueueFull = errors.New("email queue is full or timed out")

func (s *EmailService) SendVerificationEmailAsync(to, token string) error {
	select {
	case s.tasks <- EmailTask{To: to, Token: token}:
		log.Printf("Email task added to queue for %s", to)
		return nil
	case <-time.After(1 * time.Second): // 작업 추가 제한 시간
		log.Printf("Failed to add email task for %s: queue timeout", to)
		return ErrEmailQueueFull
	}
}

// 동기 이메일 전송: 실제 이메일 전송 처리
func (s *EmailService) sendEmail(to, token string) error {
	from := s.username
	auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)

	link := fmt.Sprintf("http://localhost:8080/verify-email?token=%s", token)
	body := fmt.Sprintf("Click the link to verify your email: %s", link)
	message := []byte("Subject: Email Verification\r\n\r\n" + body)

	time.Sleep(2 * time.Second) // 이메일 전송 지연 시뮬레이션 (테스트용)
	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
