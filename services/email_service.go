package services

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	gomail "gopkg.in/mail.v2"
)

// EmailService sends emails via SMTP (Yandex and compatible providers).
type EmailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewEmailService reads SMTP settings from the environment.
// Returns (nil, nil) when SMTP_USER / SMTP_PASSWORD are unset (email disabled).
//
// Required when enabled:
//   - SMTP_USER — full Yandex address (login)
//   - SMTP_PASSWORD — 16-char app password (no spaces)
//
// Optional (defaults shown):
//   - SMTP_HOST=smtp.yandex.ru
//   - SMTP_PORT=465
//   - SMTP_FROM — defaults to SMTP_USER
func NewEmailService() (*EmailService, error) {
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
	if user == "" && pass == "" {
		log.Printf("SMTP not configured (SMTP_USER / SMTP_PASSWORD); /admin/send-code disabled")
		return nil, nil
	}
	if user == "" || pass == "" {
		return nil, fmt.Errorf("both SMTP_USER and SMTP_PASSWORD must be set")
	}

	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if host == "" {
		host = "smtp.yandex.ru"
	}

	port := 465
	if raw := strings.TrimSpace(os.Getenv("SMTP_PORT")); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil || p <= 0 {
			return nil, fmt.Errorf("invalid SMTP_PORT: %q", raw)
		}
		port = p
	}

	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		from = user
	}

	return &EmailService{
		host:     host,
		port:     port,
		username: user,
		password: pass,
		from:     from,
	}, nil
}

// GenerateOTP returns a cryptographically random 6-digit code as a string.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// SendVerificationCode emails a plain-text confirmation code to the recipient.
func (s *EmailService) SendVerificationCode(to, code string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Проверочный код")
	msg.SetBody("text/plain", fmt.Sprintf("Ваш проверочный код для программы Memory Guard: %s", code))

	dialer := gomail.NewDialer(s.host, s.port, s.username, s.password)
	if s.port == 465 {
		dialer.SSL = true
	}

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}
