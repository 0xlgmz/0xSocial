package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestResendSenderSendsVerificationEmail(t *testing.T) {
	t.Parallel()

	var received resendEmail
	sender := newTestResendSender(t)
	sender.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", request.Method)
		}
		if authorization := request.Header.Get("Authorization"); authorization != "Bearer re_test" {
			t.Errorf("Authorization = %q", authorization)
		}
		if contentType := request.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type = %q", contentType)
		}
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Errorf("decode request: %v", err)
		}
		return httpResponse(http.StatusOK, `{"id":"email-id"}`), nil
	})}

	if err := sender.SendVerificationEmail(context.Background(), "person@example.com", "token/? &"); err != nil {
		t.Fatalf("SendVerificationEmail: %v", err)
	}

	if received.From != "0xSocial <noreply@example.com>" {
		t.Errorf("From = %q", received.From)
	}
	if len(received.To) != 1 || received.To[0] != "person@example.com" {
		t.Errorf("To = %#v", received.To)
	}
	if received.Subject != "Verify your 0xSocial email" {
		t.Errorf("Subject = %q", received.Subject)
	}
	if !strings.Contains(received.HTML, "Verify your email") || received.Text == "" {
		t.Error("expected non-empty verification HTML and text bodies")
	}
	if len(received.Tags) != 1 || received.Tags[0].Value != "email_verification" {
		t.Errorf("Tags = %#v", received.Tags)
	}

	verificationURL := sender.actionURL("/verify-email", "token/? &")
	parsedURL, err := url.Parse(verificationURL)
	if err != nil {
		t.Fatalf("parse verification URL: %v", err)
	}
	if parsedURL.Path != "/app/verify-email" {
		t.Errorf("path = %q", parsedURL.Path)
	}
	if token := parsedURL.Query().Get("token"); token != "token/? &" {
		t.Errorf("token = %q", token)
	}
	if !strings.Contains(received.Text, verificationURL) {
		t.Errorf("text body does not contain action URL %q", verificationURL)
	}
}

func TestResendSenderSendsPasswordEmails(t *testing.T) {
	t.Parallel()

	requests := make(chan resendEmail, 2)
	sender := newTestResendSender(t)
	sender.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var email resendEmail
		if err := json.NewDecoder(request.Body).Decode(&email); err != nil {
			t.Errorf("decode request: %v", err)
		}
		requests <- email
		return httpResponse(http.StatusOK, `{}`), nil
	})}

	if err := sender.SendPasswordResetEmail(context.Background(), "person@example.com", "reset-token"); err != nil {
		t.Fatalf("SendPasswordResetEmail: %v", err)
	}
	resetEmail := <-requests
	if !strings.Contains(resetEmail.Text, "/app/reset-password?token=reset-token") {
		t.Errorf("password reset text = %q", resetEmail.Text)
	}

	if err := sender.SendPasswordChangedEmail(context.Background(), "person@example.com"); err != nil {
		t.Fatalf("SendPasswordChangedEmail: %v", err)
	}
	changedEmail := <-requests
	if !strings.Contains(changedEmail.Text, "/app/forgot-password") {
		t.Errorf("password changed text = %q", changedEmail.Text)
	}
}

func TestResendSenderReturnsProviderError(t *testing.T) {
	t.Parallel()

	sender := newTestResendSender(t)
	sender.client = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return httpResponse(http.StatusUnauthorized, `{"message":"invalid API key"}`), nil
	})}

	err := sender.SendPasswordChangedEmail(context.Background(), "person@example.com")
	if err == nil || !strings.Contains(err.Error(), "401 Unauthorized") {
		t.Fatalf("error = %v, want provider status", err)
	}
}

func TestResendConfigFromEnv(t *testing.T) {
	for _, name := range []string{"RESEND_API_KEY", "RESEND_FROM", "APP_PUBLIC_URL"} {
		t.Setenv(name, "")
	}

	_, configured, err := ResendConfigFromEnv()
	if err != nil || configured {
		t.Fatalf("empty configuration = configured %v, error %v", configured, err)
	}

	t.Setenv("RESEND_API_KEY", "re_test")
	if _, configured, err = ResendConfigFromEnv(); err == nil || configured {
		t.Fatalf("partial configuration = configured %v, error %v", configured, err)
	}

	t.Setenv("RESEND_FROM", "0xSocial <noreply@example.com>")
	t.Setenv("APP_PUBLIC_URL", "https://social.example")
	configuration, configured, err := ResendConfigFromEnv()
	if err != nil || !configured {
		t.Fatalf("complete configuration = configured %v, error %v", configured, err)
	}
	if configuration.APIKey != "re_test" {
		t.Errorf("APIKey = %q", configuration.APIKey)
	}
}

func TestNewResendSenderValidatesConfiguration(t *testing.T) {
	t.Parallel()

	_, err := NewResendSender(ResendConfig{
		APIKey:       "re_test",
		From:         "not-an-address",
		PublicAppURL: "https://social.example",
	})
	if err == nil {
		t.Error("expected invalid From address to fail")
	}

	_, err = NewResendSender(ResendConfig{
		APIKey:       "re_test",
		From:         "noreply@example.com",
		PublicAppURL: "social.example",
	})
	if err == nil {
		t.Error("expected relative public app URL to fail")
	}
}

func newTestResendSender(t *testing.T) *ResendSender {
	t.Helper()
	sender, err := NewResendSender(ResendConfig{
		APIKey:       "re_test",
		From:         "0xSocial <noreply@example.com>",
		PublicAppURL: "https://social.example/app/",
	})
	if err != nil {
		t.Fatalf("NewResendSender: %v", err)
	}
	return sender
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func httpResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
