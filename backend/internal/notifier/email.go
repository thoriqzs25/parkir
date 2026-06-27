package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	ToAddresses  []string
}

type Notifier struct {
	config Config
}

func New(config Config) *Notifier {
	return &Notifier{config: config}
}

func (n *Notifier) SendAlertEmail(code, locationName, description string) error {
	if n.config.SMTPHost == "" || len(n.config.ToAddresses) == 0 {
		return nil
	}

	subject := fmt.Sprintf("[PARKIR Alert] %s - %s", code, locationName)
	body := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; padding: 20px;">
<h2 style="color: #dc2626;">PARKIR Alert Triggered</h2>
<table style="border-collapse: collapse; width: 100%%;">
<tr><td style="padding: 8px; border: 1px solid #ddd; font-weight: bold;">Alert Code</td><td style="padding: 8px; border: 1px solid #ddd;">%s</td></tr>
<tr><td style="padding: 8px; border: 1px solid #ddd; font-weight: bold;">Location</td><td style="padding: 8px; border: 1px solid #ddd;">%s</td></tr>
<tr><td style="padding: 8px; border: 1px solid #ddd; font-weight: bold;">Description</td><td style="padding: 8px; border: 1px solid #ddd;">%s</td></tr>
<tr><td style="padding: 8px; border: 1px solid #ddd; font-weight: bold;">Time</td><td style="padding: 8px; border: 1px solid #ddd;">%%s</td></tr>
</table>
<p style="margin-top: 20px; color: #666;">Please log in to the PARKIR dashboard to review and acknowledge this alert.</p>
</body>
</html>`, code, locationName, description)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		n.config.FromAddress,
		strings.Join(n.config.ToAddresses, ","),
		subject,
		body)

	addr := fmt.Sprintf("%s:%s", n.config.SMTPHost, n.config.SMTPPort)
	auth := smtp.PlainAuth("", n.config.SMTPUser, n.config.SMTPPassword, n.config.SMTPHost)

	return smtp.SendMail(addr, auth, n.config.FromAddress, n.config.ToAddresses, []byte(msg))
}