package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridClient struct {
	client   *sendgrid.Client
	fromName string
	fromAddr string
}

func NewSendGridClient() *SendGridClient {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromAddr := os.Getenv("SENDGRID_FROM_EMAIL")
	fromName := os.Getenv("SENDGRID_FROM_NAME")

	if apiKey == "" || fromAddr == "" {
		log.Println("WARN: SendGrid not fully configured")
	}
	if fromName == "" {
		fromName = "Game Rental"
	}

	return &SendGridClient{
		client:   sendgrid.NewSendClient(apiKey),
		fromName: fromName,
		fromAddr: fromAddr,
	}
}

func (s *SendGridClient) SendEmail(ctx context.Context, to, subject, plainText, htmlContent string) error {
	if s.client == nil || s.fromAddr == "" {
		return fmt.Errorf("sendgrid not configured")
	}

	// Fallback plainText from HTML if empty to avoid spam marking
	if plainText == "" && htmlContent != "" {
		plainText = stripHTML(htmlContent)
	}

	from := mail.NewEmail(s.fromName, s.fromAddr)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, plainText, htmlContent)

	// Note: SendGrid client doesn't support context timeout directly
	// Consider implementing custom HTTP client wrapper if strict timeout control needed

	resp, err := s.client.Send(message)
	if err != nil {
		log.Printf("ERROR: SendGrid send failed to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	if resp.StatusCode >= 400 {
		log.Printf("ERROR: SendGrid error %d: %s", resp.StatusCode, resp.Body)
		return fmt.Errorf("sendgrid error: status=%d body=%s", resp.StatusCode, resp.Body)
	}
	log.Printf("INFO: Email sent to %s (subject: %s)", to, subject)
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
		log.Printf("ERROR: SendGrid template send failed to %s: %v", to, err)
		return fmt.Errorf("failed to send template email: %w", err)
	}
	if resp.StatusCode >= 400 {
		log.Printf("ERROR: SendGrid template error %d: %s", resp.StatusCode, resp.Body)
		return fmt.Errorf("sendgrid template error: status=%d", resp.StatusCode)
	}
	log.Printf("INFO: Template email sent to %s (template: %s)", to, templateID)
	return nil
}

// stripHTML removes HTML tags for plaintext fallback
func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	plain := re.ReplaceAllString(html, "")
	return strings.TrimSpace(plain)
}
