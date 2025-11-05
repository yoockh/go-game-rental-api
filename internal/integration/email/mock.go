package integration

import "context"

// MockEmailSender for testing
type MockEmailSender struct {
	SentEmails []MockEmail
}

type MockEmail struct {
	To          string
	Subject     string
	PlainText   string
	HTMLContent string
	TemplateID  string
	Data        map[string]interface{}
}

func (m *MockEmailSender) SendEmail(ctx context.Context, to, subject, plainText, htmlContent string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:          to,
		Subject:     subject,
		PlainText:   plainText,
		HTMLContent: htmlContent,
	})
	return nil
}
