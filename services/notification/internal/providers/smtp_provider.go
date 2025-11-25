package providers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"backend/services/notification/internal/config"
)

// SMTPProvider handles email sending via SMTP
type SMTPProvider struct {
	cfg *config.Config
}

func NewSMTPProvider(cfg *config.Config) *SMTPProvider {
	return &SMTPProvider{cfg: cfg}
}

func (p *SMTPProvider) SendEmail(to string, subject string, body string) error {
	// Setup authentication
	auth := smtp.PlainAuth("", p.cfg.SMTPUser, p.cfg.SMTPPassword, p.cfg.SMTPHost)

	// Compose message
	msg := []byte(fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// Connect to server
	addr := fmt.Sprintf("%s:%d", p.cfg.SMTPHost, p.cfg.SMTPPort)
	
	if p.cfg.SMTPTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         p.cfg.SMTPHost,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %v", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, p.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %v", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %v", err)
		}

		if err = client.Mail(p.cfg.SMTPUser); err != nil {
			return fmt.Errorf("failed to set sender: %v", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %v", err)
		}

		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to open data writer: %v", err)
		}

		if _, err = writer.Write(msg); err != nil {
			return fmt.Errorf("failed to write message: %v", err)
		}

		if err = writer.Close(); err != nil {
			return fmt.Errorf("failed to close data writer: %v", err)
		}

		if err = client.Quit(); err != nil {
			return fmt.Errorf("failed to quit: %v", err)
		}
	} else {
		// Plain SMTP
		err := smtp.SendMail(addr, auth, p.cfg.SMTPUser, []string{to}, msg)
		if err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

