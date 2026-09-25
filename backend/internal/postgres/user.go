package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

func ReplacePasswordResetToken(ctx context.Context, pool *pgxpool.Pool, emailNormalized string, tokenHash []byte) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var userID int64

	err = tx.QueryRow(ctx, `
		SELECT id
		FROM users
		WHERE email_normalized = $1
		  AND status = 'active'
		FOR UPDATE
	`, emailNormalized).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth_tokens
		SET used_at = NOW()
		WHERE user_id = $1
		  AND purpose = 'reset_password'
		  AND used_at IS NULL
	`, userID)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO auth_tokens (
			user_id,
			purpose,
			token_hash,
			expires_at
		)
		VALUES (
			$1,
			'reset_password',
			$2,
			NOW() + INTERVAL '30 minutes'
		)
	`, userID, tokenHash)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}
func ResetPassword(ctx context.Context, pool *pgxpool.Pool, rawToken string, passwordHash string) (string, error) {
	tokenHash := auth.HashToken(rawToken)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var userID int64

	err = tx.QueryRow(ctx, `
		UPDATE auth_tokens
		SET used_at = NOW()
		WHERE token_hash = $1
		  AND purpose = 'reset_password'
		  AND used_at IS NULL
		  AND expires_at > NOW()
		RETURNING user_id
	`, tokenHash).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidPasswordResetToken
	}
	if err != nil {
		return "", err
	}

	result, err := tx.Exec(ctx, `
		UPDATE password_credentials
		SET password_hash = $1,
		    password_changed_at = NOW()
		WHERE user_id = $2
	`, passwordHash, userID)
	if err != nil {
		return "", err
	}
	if result.RowsAffected() != 1 {
		return "", errors.New("password credential not found")
	}

	_, err = tx.Exec(ctx, `
		UPDATE auth_tokens
		SET used_at = NOW()
		WHERE user_id = $1
		  AND purpose = 'reset_password'
		  AND used_at IS NULL
	`, userID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = NOW(),
		    revoked_reason = 'password_reset'
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		UPDATE users
		SET sessions_valid_after = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return "", err
	}

	var email string

	err = tx.QueryRow(
		ctx,
		`UPDATE users
		SET sessions_valid_after = NOW(),
			updated_at = NOW()
		WHERE id = $1
		RETURNING email`,
		userID,
	).Scan(&email)

	return email, tx.Commit(ctx)
}

func ListUserSessions(ctx context.Context, pool *pgxpool.Pool, userID int64) ([]UserSession, error) {
	rows, err := pool.Query(
		ctx,
		`SELECT
			id,
			COALESCE(user_agent, ''),
			COALESCE(HOST(ip_address), ''),
			auth_method,
			risk_level,
			created_at,
			last_seen_at,
			expires_at
		 FROM sessions
		 WHERE user_id = $1
		   AND revoked_at IS NULL
		   AND expires_at > NOW()
		 ORDER BY last_seen_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]UserSession, 0)

	for rows.Next() {
		var session UserSession

		err := rows.Scan(
			&session.ID,
			&session.UserAgent,
			&session.IPAddress,
			&session.AuthMethod,
			&session.RiskLevel,
			&session.CreatedAt,
			&session.LastSeenAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user session: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read user sessions: %w", err)
	}

	return sessions, nil
}
