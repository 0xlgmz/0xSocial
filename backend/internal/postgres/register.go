package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

func InsertRegisteredUser(ctx context.Context, pool *pgxpool.Pool, emailNormalized, hashedPassword string, verificationTokenHash []byte) error {
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

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email verification: %w", err)
	}

	return nil
}
