package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"app4every/services/auth/internal/config"
)

type MailerService interface {
	SendPasswordResetEmail(toEmail, token string) error
	SendVerificationEmail(toEmail, token string) error
}

type mailerService struct {
	cfg *config.Config
}

func NewMailerService(cfg *config.Config) MailerService {
	return &mailerService{cfg: cfg}
}

func (s *mailerService) sendMail(to, subject, body string) error {
	from := s.cfg.SMTPUser
	pass := s.cfg.SMTPPassword
	host := s.cfg.SMTPHost
	port := s.cfg.SMTPPort

	if host == "" || port == "" || from == "" || pass == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		body

	auth := smtp.PlainAuth("", from, pass, host)

	// In a real production system, you'd want to use TLS properly
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", host+":"+port, tlsconfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

func (s *mailerService) SendPasswordResetEmail(toEmail, token string) error {
	// Construct the frontend reset link (assuming frontend runs on localhost:3000 for now, or configured)
	// We'll just hardcode it or pass it via config
	frontendURL := "http://localhost:3000" // TODO: move to config
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	subject := "Сброс пароля - App4Every"
	body := fmt.Sprintf(`
		<h1>Сброс пароля</h1>
		<p>Вы запросили сброс пароля. Перейдите по ссылке ниже, чтобы установить новый пароль:</p>
		<a href="%s">%s</a>
		<p>Ссылка действительна 1 час.</p>
	`, resetLink, resetLink)

	return s.sendMail(toEmail, subject, body)
}

func (s *mailerService) SendVerificationEmail(toEmail, token string) error {
	frontendURL := "http://localhost:3000"
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	subject := "Подтверждение почты - App4Every"
	body := fmt.Sprintf(`
		<h1>Подтверждение почты</h1>
		<p>Перейдите по ссылке ниже, чтобы подтвердить вашу почту:</p>
		<a href="%s">%s</a>
		<p>Ссылка действительна 24 часа.</p>
	`, verifyLink, verifyLink)

	return s.sendMail(toEmail, subject, body)
}
