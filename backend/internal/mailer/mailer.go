package mailer

import "context"

type Sender interface {
	SendVerificationEmail(
		ctx context.Context,
		email string,
		rawToken string,
	) error
}
