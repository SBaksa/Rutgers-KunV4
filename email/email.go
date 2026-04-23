package email

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// LoadSMTPConfig loads SMTP settings from environment variables
func LoadSMTPConfig() *SMTPConfig {
	return &SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),     // e.g., "smtp.gmail.com"
		Port:     os.Getenv("SMTP_PORT"),     // e.g., "587"
		Username: os.Getenv("SMTP_USERNAME"), // your email
		Password: os.Getenv("SMTP_PASSWORD"), // app password
		From:     os.Getenv("SMTP_FROM"),     // sender email
	}
}

// IsConfigured checks if SMTP is properly set up
func (c *SMTPConfig) IsConfigured() bool {
	return c.Host != "" && c.Port != "" && c.Username != "" && c.Password != "" && c.From != ""
}

// GenerateVerificationCode creates a 6-digit verification code
func GenerateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

// SendVerificationEmail sends verification code to Rutgers email
func (c *SMTPConfig) SendVerificationEmail(netID, verificationCode, serverName string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	to := fmt.Sprintf("%s@scarletmail.rutgers.edu", netID)
	subject := fmt.Sprintf("Verify your role in %s!", serverName)

	body := fmt.Sprintf(`Your verification code is:

%s

Enter this code in Discord to verify your Rutgers email and get access to the server.

This code will expire in 15 minutes.`, verificationCode)

	return c.sendEmail(to, subject, body)
}

// sendEmail handles the actual SMTP sending
func (c *SMTPConfig) sendEmail(to, subject, body string) error {
	// Construct the email message
	message := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body)

	// Set up authentication
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)

	// Send the email
	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)
	err := smtp.SendMail(addr, auth, c.From, []string{to}, []byte(message))

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
