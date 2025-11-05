package email

import "context"

// EmailSender interface for email service abstraction
type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, plainText, htmlContent string) error
	SendWithTemplate(ctx context.Context, to, templateID string, dynamicData map[string]interface{}) error
}
