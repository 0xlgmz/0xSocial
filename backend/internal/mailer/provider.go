package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"
)

const resendEmailEndpoint = "https://api.resend.com/emails"

type ResendConfig struct {
	APIKey       string
	From         string
	PublicAppURL string
}

type ResendSender struct {
	apiKey       string
	from         string
	publicAppURL *url.URL
	client       *http.Client
	endpoint     string
}

type resendEmail struct {
	From    string           `json:"from"`
	To      []string         `json:"to"`
	Subject string           `json:"subject"`
	HTML    string           `json:"html"`
	Text    string           `json:"text"`
	Tags    []resendEmailTag `json:"tags,omitempty"`
}

type resendEmailTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type emailTemplateData struct {
	ActionURL string
}

var _ Sender = (*ResendSender)(nil)

var actionEmailTemplate = template.Must(template.New("action-email").Parse(`<!doctype html>
<html lang="en">
<body style="margin:0;background:#f4f4f5;font-family:Arial,sans-serif;color:#18181b">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding:32px 16px">
    <tr><td align="center">
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border-radius:12px;padding:32px">
        <tr><td>
          <p style="margin:0 0 24px;font-size:20px;font-weight:700">0xSocial</p>
          <h1 style="margin:0 0 16px;font-size:24px">{{.Heading}}</h1>
          <p style="margin:0 0 24px;line-height:1.6">{{.Body}}</p>
          <p style="margin:0 0 24px"><a href="{{.ActionURL}}" style="display:inline-block;background:#18181b;color:#ffffff;text-decoration:none;padding:12px 20px;border-radius:8px;font-weight:700">{{.ActionLabel}}</a></p>
          <p style="margin:0 0 8px;font-size:13px;color:#71717a">If the button does not work, copy and paste this link:</p>
          <p style="margin:0;word-break:break-all;font-size:13px"><a href="{{.ActionURL}}">{{.ActionURL}}</a></p>
          <p style="margin:24px 0 0;font-size:13px;color:#71717a">{{.Expiry}} If you did not request this, you can ignore this email.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`))

var notificationEmailTemplate = template.Must(template.New("notification-email").Parse(`<!doctype html>
<html lang="en">
<body style="margin:0;background:#f4f4f5;font-family:Arial,sans-serif;color:#18181b">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding:32px 16px">
    <tr><td align="center">
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border-radius:12px;padding:32px">
        <tr><td>
          <p style="margin:0 0 24px;font-size:20px;font-weight:700">0xSocial</p>
          <h1 style="margin:0 0 16px;font-size:24px">Your password was changed</h1>
          <p style="margin:0 0 24px;line-height:1.6">The password for your 0xSocial account was changed successfully. All existing sessions have been signed out.</p>
          <p style="margin:0;line-height:1.6">If you did not make this change, reset your password immediately at <a href="{{.ActionURL}}">{{.ActionURL}}</a>.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`))

func ResendConfigFromEnv() (ResendConfig, bool, error) {
	configuration := ResendConfig{
		APIKey:       strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		From:         strings.TrimSpace(os.Getenv("RESEND_FROM")),
		PublicAppURL: strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")),
	}

	if configuration.APIKey == "" && configuration.From == "" && configuration.PublicAppURL == "" {
		return ResendConfig{}, false, nil
	}

	missing := make([]string, 0, 3)
	for name, value := range map[string]string{
		"RESEND_API_KEY": configuration.APIKey,
		"RESEND_FROM":    configuration.From,
		"APP_PUBLIC_URL": configuration.PublicAppURL,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return ResendConfig{}, false, fmt.Errorf(
			"incomplete Resend configuration: missing %s",
			strings.Join(missing, ", "),
		)
	}

	return configuration, true, nil
}

func NewResendSender(configuration ResendConfig) (*ResendSender, error) {
	if strings.TrimSpace(configuration.APIKey) == "" {
		return nil, errors.New("Resend API key is required")
	}
	if strings.ContainsAny(configuration.From, "\r\n") {
		return nil, errors.New("RESEND_FROM must not contain newlines")
	}
	_, err := mail.ParseAddress(configuration.From)
	if err != nil {
		return nil, fmt.Errorf("RESEND_FROM must be a valid email address: %w", err)
	}

	publicAppURL, err := url.Parse(configuration.PublicAppURL)
	if err != nil || publicAppURL.Scheme == "" || publicAppURL.Host == "" {
		return nil, errors.New("APP_PUBLIC_URL must be an absolute URL")
	}
	if publicAppURL.Scheme != "http" && publicAppURL.Scheme != "https" {
		return nil, errors.New("APP_PUBLIC_URL must use http or https")
	}
	publicAppURL.RawQuery = ""
	publicAppURL.Fragment = ""
	publicAppURL.Path = strings.TrimRight(publicAppURL.Path, "/")

	return &ResendSender{
		apiKey:       strings.TrimSpace(configuration.APIKey),
		from:         strings.TrimSpace(configuration.From),
		publicAppURL: publicAppURL,
		client:       &http.Client{Timeout: 10 * time.Second},
		endpoint:     resendEmailEndpoint,
	}, nil
}

func (sender *ResendSender) SendVerificationEmail(ctx context.Context, email string, rawToken string) error {
	actionURL := sender.actionURL("/verify-email", rawToken)
	html, err := renderActionEmail(actionURL, "Verify your email", "Confirm your email address to finish creating your 0xSocial account.", "Verify email", "This link expires in 8 hours.")
	if err != nil {
		return fmt.Errorf("render verification email: %w", err)
	}

	return sender.send(ctx, resendEmail{
		From:    sender.from,
		To:      []string{email},
		Subject: "Verify your 0xSocial email",
		HTML:    html,
		Text:    fmt.Sprintf("Verify your email to finish creating your 0xSocial account:\n\n%s\n\nThis link expires in 8 hours. If you did not request this, you can ignore this email.", actionURL),
		Tags:    []resendEmailTag{{Name: "category", Value: "email_verification"}},
	})
}

func (sender *ResendSender) SendPasswordResetEmail(ctx context.Context, email string, rawToken string) error {
	actionURL := sender.actionURL("/reset-password", rawToken)
	html, err := renderActionEmail(actionURL, "Reset your password", "Use the link below to choose a new password for your 0xSocial account.", "Reset password", "This link expires in 30 minutes.")
	if err != nil {
		return fmt.Errorf("render password reset email: %w", err)
	}

	return sender.send(ctx, resendEmail{
		From:    sender.from,
		To:      []string{email},
		Subject: "Reset your 0xSocial password",
		HTML:    html,
		Text:    fmt.Sprintf("Reset your 0xSocial password:\n\n%s\n\nThis link expires in 30 minutes. If you did not request this, you can ignore this email.", actionURL),
		Tags:    []resendEmailTag{{Name: "category", Value: "password_reset"}},
	})
}

func (sender *ResendSender) SendPasswordChangedEmail(ctx context.Context, email string) error {
	actionURL := sender.actionURL("/forgot-password", "")
	var html bytes.Buffer
	if err := notificationEmailTemplate.Execute(&html, emailTemplateData{ActionURL: actionURL}); err != nil {
		return fmt.Errorf("render password changed email: %w", err)
	}

	return sender.send(ctx, resendEmail{
		From:    sender.from,
		To:      []string{email},
		Subject: "Your 0xSocial password was changed",
		HTML:    html.String(),
		Text:    fmt.Sprintf("Your 0xSocial password was changed successfully, and all existing sessions were signed out.\n\nIf you did not make this change, reset your password immediately: %s", actionURL),
		Tags:    []resendEmailTag{{Name: "category", Value: "password_changed"}},
	})
}

func (sender *ResendSender) actionURL(path, rawToken string) string {
	actionURL := *sender.publicAppURL
	actionURL.Path = strings.TrimRight(actionURL.Path, "/") + path
	if rawToken != "" {
		query := actionURL.Query()
		query.Set("token", rawToken)
		actionURL.RawQuery = query.Encode()
	}
	return actionURL.String()
}

func (sender *ResendSender) send(ctx context.Context, email resendEmail) error {
	body, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("encode Resend email: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Resend request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+sender.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		return fmt.Errorf("send email with Resend: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 4_096))
		if readErr != nil {
			return fmt.Errorf("Resend returned %s", response.Status)
		}
		message := strings.TrimSpace(string(responseBody))
		if message == "" {
			message = "empty response"
		}
		return fmt.Errorf("Resend returned %s: %s", response.Status, message)
	}

	return nil
}

func renderActionEmail(actionURL, heading, body, actionLabel, expiry string) (string, error) {
	var rendered bytes.Buffer
	err := actionEmailTemplate.Execute(&rendered, struct {
		ActionURL   string
		Heading     string
		Body        string
		ActionLabel string
		Expiry      string
	}{
		ActionURL:   actionURL,
		Heading:     heading,
		Body:        body,
		ActionLabel: actionLabel,
		Expiry:      expiry,
	})
	return rendered.String(), err
}
