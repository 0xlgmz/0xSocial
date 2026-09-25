package mailer

import (
	"context"
	"log/slog"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (sender *LogSender) SendVerificationEmail(_ context.Context, email string, rawToken string) error {
	slog.Info(
		"development verification email",
		"email", email,
		"token", rawToken,
	)

	return nil
}

func (sender *LogSender) SendPasswordResetEmail(_ context.Context, email string, rawToken string) error {
	slog.Info(
		"development password reset email",
		"email", email,
		"token", rawToken,
	)

	return nil
}

func (sender *LogSender) SendPasswordChangedEmail(_ context.Context, email string) error {
	slog.Info(
		"development password changed email",
		"email", email,
	)

	return nil
}
