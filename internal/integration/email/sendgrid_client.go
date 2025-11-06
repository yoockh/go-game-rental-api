package email

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

type SendGridClient struct {
	client   *sendgrid.Client
	fromName string
	fromAddr string
}

func NewSendGridClient() (*SendGridClient, error) {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromAddr := os.Getenv("SENDGRID_FROM_EMAIL")
	fromName := os.Getenv("SENDGRID_FROM_NAME")

	if apiKey == "" || fromAddr == "" {
		return nil, fmt.Errorf("sendgrid not configured: missing API_KEY or FROM_EMAIL")
	}
	if fromName == "" {
		fromName = "Game Rental"
	}

	return &SendGridClient{
		client:   sendgrid.NewSendClient(apiKey),
		fromName: fromName,
		fromAddr: fromAddr,
	}, nil
}

func (s *SendGridClient) SendEmail(ctx context.Context, to, subject, plainText, htmlContent string) error {
	if !isValidEmail(to) {
		return fmt.Errorf("invalid email address: %s", to)
	}

	// Fallback plainText from HTML if empty to avoid spam marking
	if plainText == "" && htmlContent != "" {
		plainText = stripHTML(htmlContent)
	}

	from := mail.NewEmail(s.fromName, s.fromAddr)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, plainText, htmlContent)

	resp, err := s.client.Send(message)
	if err != nil {
		logrus.WithError(err).WithField("to", to).Error("SendGrid send failed")
		return fmt.Errorf("failed to send email: %w", err)
	}
	if resp.StatusCode >= 400 {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   resp.Body,
			"to":     to,
		}).Error("SendGrid error")
		return fmt.Errorf("sendgrid error: status=%d", resp.StatusCode)
	}
	logrus.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email sent successfully")
	return nil
}

func (s *SendGridClient) SendWithTemplate(ctx context.Context, to, templateID string, dynamicData map[string]interface{}) error {
	if s.client == nil || s.fromAddr == "" {
		return fmt.Errorf("sendgrid not configured")
	}
	from := mail.NewEmail(s.fromName, s.fromAddr)
	toEmail := mail.NewEmail("", to)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.SetTemplateID(templateID)

	p := mail.NewPersonalization()
	p.AddTos(toEmail)
	for k, v := range dynamicData {
		p.SetDynamicTemplateData(k, v)
	}
	message.AddPersonalizations(p)

	resp, err := s.client.Send(message)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"to":         to,
			"template_id": templateID,
		}).Error("SendGrid template send failed")
		return fmt.Errorf("failed to send template email: %w", err)
	}
	if resp.StatusCode >= 400 {
		logrus.WithFields(logrus.Fields{
			"status":     resp.StatusCode,
			"to":         to,
			"template_id": templateID,
		}).Error("SendGrid template error")
		return fmt.Errorf("sendgrid template error: status=%d", resp.StatusCode)
	}
	logrus.WithFields(logrus.Fields{
		"to":         to,
		"template_id": templateID,
	}).Info("Template email sent successfully")
	return nil
}

// stripHTML removes HTML tags for plaintext fallback
func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	plain := re.ReplaceAllString(html, "")
	return strings.TrimSpace(plain)
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
