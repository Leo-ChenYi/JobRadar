package notifier

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"jobradar/internal/config"
	"jobradar/internal/model"
)

// EmailNotifier sends notifications via email
type EmailNotifier struct {
	config config.EmailConfig
}

// NewEmail creates a new Email notifier
func NewEmail(cfg config.EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		config: cfg,
	}
}

// Name returns the notifier name
func (e *EmailNotifier) Name() string {
	return "email"
}

// Send sends a notification for a matched job
func (e *EmailNotifier) Send(matched *model.MatchedJob) error {
	subject := FormatEmailSubject(matched)
	body := FormatEmailBody(matched)
	return e.sendEmail(subject, body)
}

// SendTest sends a test notification
func (e *EmailNotifier) SendTest() error {
	subject := "[JobRadar] Test Notification"
	body := `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body>
    <h2>🔔 JobRadar Test Notification</h2>
    <p>This is a test message to verify your email notification settings are working correctly.</p>
    <p>If you received this email, your configuration is correct!</p>
    <ul>
        <li>✅ SMTP Connection: Successful</li>
        <li>✅ Authentication: Verified</li>
        <li>✅ Email Delivery: Working</li>
    </ul>
</body>
</html>`
	return e.sendEmail(subject, body)
}

// sendEmail sends an email with the given subject and HTML body
func (e *EmailNotifier) sendEmail(subject, body string) error {
	from := e.config.Username
	to := e.config.To

	// Build email headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	// Use TLS for port 465, STARTTLS for other ports
	if e.config.SMTPPort == 465 {
		return e.sendWithTLS(addr, from, to, msg.String())
	}
	return e.sendWithSTARTTLS(addr, from, to, msg.String())
}

// sendWithSTARTTLS sends email using STARTTLS (port 587)
func (e *EmailNotifier) sendWithSTARTTLS(addr, from, to, msg string) error {
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)

	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendWithTLS sends email using direct TLS (port 465)
func (e *EmailNotifier) sendWithTLS(addr, from, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: e.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}
