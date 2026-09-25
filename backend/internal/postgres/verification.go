package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidVerificationToken = errors.New(
	"invalid or expired verification token",
)

func ReplaceEmailVerificationToken(ctx context.Context, pool *pgxpool.Pool, emailNormalized string, tokenHash []byte, ipAddress string, userAgent string) (string, bool, error) {
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

func VerifyEmailAndCreateSession(ctx context.Context, pool *pgxpool.Pool, rawVerificationToken string, sessionTokenHash []byte, ipAddress string, userAgent string) (time.Time, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("begin verification transaction: %w", err)
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
		auth.HashToken(rawVerificationToken),
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrInvalidVerificationToken
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("consume verification token: %w", err)
	}

	result, err := tx.Exec(
		ctx,
		`UPDATE users
		 SET email_verified_at = NOW(),
		     status = 'active',
		     updated_at = NOW()
		 WHERE id = $1
		   AND status = 'pending_verification'`,
		userID,
	)
	if err != nil {
		return time.Time{}, fmt.Errorf("activate user: %w", err)
	}

	if result.RowsAffected() != 1 {
		return time.Time{}, ErrInvalidVerificationToken
	}

	var (
		sessionID int64
		expiresAt time.Time
	)

	err = tx.QueryRow(
		ctx,
		`INSERT INTO sessions (
			user_id,
			token_hash,
			ip_address,
			user_agent,
			auth_method,
			expires_at
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, '')::INET,
			$4,
			'email_verification',
			NOW() + INTERVAL '7 days'
		)
		RETURNING id, expires_at`,
		userID,
		sessionTokenHash,
		ipAddress,
		userAgent,
	).Scan(&sessionID, &expiresAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("create verification session: %w", err)
	}

	event := AuthEvent{
		UserID:    &userID,
		SessionID: &sessionID,
		EventType: AuthEventEmailVerified,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := InsertAuthEvent(ctx, tx, event); err != nil {
		return time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return time.Time{}, fmt.Errorf("commit verification: %w", err)
	}

	return expiresAt, nil
}
