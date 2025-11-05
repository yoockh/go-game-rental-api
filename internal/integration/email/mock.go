package email

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

func (m *MockEmailSender) SendWithTemplate(ctx context.Context, to, templateID string, dynamicData map[string]interface{}) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:         to,
		TemplateID: templateID,
		Data:       dynamicData,
	})
	return nil
}
