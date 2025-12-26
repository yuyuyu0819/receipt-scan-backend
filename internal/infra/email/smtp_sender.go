package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

// SMTPSender sends email via an SMTP server.
type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSMTPSenderFromEnv builds an SMTP sender using environment variables.
func NewSMTPSenderFromEnv() (*SMTPSender, error) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if host == "" || port == "" || username == "" || password == "" || from == "" {
		return nil, errors.New("SMTP_HOST/SMTP_PORT/SMTP_USER/SMTP_PASS/SMTP_FROM must be set")
	}

	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}, nil
}

// SendConfirmation sends a basic confirmation email.
func (s *SMTPSender) SendConfirmation(ctx context.Context, address string, userID int64) error {
	_ = ctx
	subject := "Receipt Scan: Account Confirmation"
	activationURL := fmt.Sprintf("http://localhost:8080/activate?userId=%d", userID)
	body := fmt.Sprintf("User %d, your account has been created.\nActivate here: %s\n", userID, activationURL)
	msg := []byte("To: " + address + "\r\n" +
		"From: " + s.from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err := smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{address}, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
