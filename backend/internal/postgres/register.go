package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrHandleAlreadyExists = errors.New("handle already exists")

func InsertRegisteredUser(ctx context.Context, pool *pgxpool.Pool, emailNormalized, userHandle, hashedPassword string, verificationTokenHash []byte, ipAddress, userAgent string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errors.New("could not create account")
	}
	defer tx.Rollback(ctx)

	var userID int64

	err = tx.QueryRow(
		ctx,
		`INSERT INTO users (email, email_normalized)
		VALUES ($1, $2)
		ON CONFLICT (email_normalized) DO NOTHING
		RETURNING id`,
		emailNormalized,
		emailNormalized,
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEmailAlreadyExists
	}

	if err != nil {
		return errors.New("could not create account")
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO password_credentials (user_id, password_hash)
		VALUES ($1, $2)`,
		userID,
		hashedPassword,
	)
	if err != nil {
		return errors.New("could not create account")
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO profiles (user_id, handle)
		VALUES ($1, $2)`,
		userID,
		userHandle,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "profiles_handle_key" {
			return ErrHandleAlreadyExists
		}

		return fmt.Errorf("create user profile: %w", err)
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
		verificationTokenHash,
	)
	if err != nil {
		return fmt.Errorf("insert verification token: %w", err)
	}

	event := AuthEvent{
		UserID:    &userID,
		EventType: AuthEventRegistrationSucceeded,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := InsertAuthEvent(ctx, tx, event); err != nil {
		return fmt.Errorf("record registration event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email verification: %w", err)
	}

	return nil
}
