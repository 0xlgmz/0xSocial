package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidVerificationToken = errors.New(
	"invalid or expired verification token",
)

func VerifyEmailToken(ctx context.Context, pool *pgxpool.Pool, rawToken string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errors.New("email validation failed")
	}
	defer tx.Rollback(ctx)

	var userID int64

	err = tx.QueryRow(
		ctx,
		`UPDATE auth_tokens
		SET used_at = NOW()
		WHERE token_hash = $1
			AND purpose = 'verify_email'
			AND used_at IS NULL
			AND expires_at > NOW()
		RETURNING user_id`,
		auth.HashToken(rawToken),
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidVerificationToken
	}
	if err != nil {
		return fmt.Errorf("consume verification token: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE users
		SET email_verified_at = NOW(),
			status = 'active',
			updated_at = NOW()
		WHERE id = $1
		AND status = 'pending_verification'`,
		userID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidVerificationToken
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email verification: %w", err)
	}

	return nil
}

func ReplaceEmailVerificationToken(ctx context.Context, pool *pgxpool.Pool, emailNormalized string, tokenHash []byte) (string, bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("begin resend transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		userID int64
		email  string
	)

	err = tx.QueryRow(
		ctx,
		`SELECT id, email
		 FROM users
		 WHERE email_normalized = $1
		   AND status = 'pending_verification'
		 FOR UPDATE`,
		emailNormalized,
	).Scan(&userID, &email)

	if errors.Is(err, pgx.ErrNoRows) {
		// Return success-like behavior to prevent account enumeration.
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("find pending user: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE auth_tokens
		 SET used_at = NOW()
		 WHERE user_id = $1
		   AND purpose = 'verify_email'
		   AND used_at IS NULL`,
		userID,
	)

	if err != nil {
		return "", false, fmt.Errorf("invalidate verification tokens: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO auth_tokens (
		     user_id,
		     purpose,
		     token_hash,
		     expires_at
		 )
		 VALUES (
		     $1,
		     'verify_email',
		     $2,
		     NOW() + INTERVAL '8 hours'
		 )`,
		userID,
		tokenHash,
	)

	if err != nil {
		return "", false, fmt.Errorf("insert verification token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", false, fmt.Errorf("commit resend transaction: %w", err)
	}

	return email, true, nil
}
