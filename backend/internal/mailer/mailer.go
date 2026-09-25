package mailer

import "context"

type Sender interface {
	SendVerificationEmail(
		ctx context.Context,
		email string,
		rawToken string,
	) error

	SendPasswordResetEmail(
		ctx context.Context,
		email string,
		rawToken string,
	) error

	SendPasswordChangedEmail(
		ctx context.Context,
		email string,
	) error
}
