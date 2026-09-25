package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSession struct {
	ID         int64
	UserAgent  string
	IPAddress  string
	AuthMethod string
	RiskLevel  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
}

func LogoutSession(ctx context.Context, pool *pgxpool.Pool, tokenHash []byte, ipAddress, userAgent string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin logout transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		sessionID int64
		userID    int64
	)

	err = tx.QueryRow(
		ctx,
		`UPDATE sessions
		 SET revoked_at = NOW(),
		     revoked_reason = 'user_logout'
		 WHERE token_hash = $1
		   AND revoked_at IS NULL
		 RETURNING id, user_id`,
		tokenHash,
	).Scan(&sessionID, &userID)

	if errors.Is(err, pgx.ErrNoRows) {
		// Missing or already-revoked session is still a successful logout.
		return nil
	}
	if err != nil {
		return fmt.Errorf("revoke logout session: %w", err)
	}

	event := AuthEvent{
		UserID:    &userID,
		SessionID: &sessionID,
		EventType: AuthEventLogout,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := InsertAuthEvent(ctx, tx, event); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit logout: %w", err)
	}

	return nil
}

func RevokeUserSession(ctx context.Context, pool *pgxpool.Pool, userID, targetSessionID, currentSessionID int64, ipAddress, userAgent string) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin session revocation: %w", err)
	}
	defer tx.Rollback(ctx)

	var revokedSessionID int64

	err = tx.QueryRow(
		ctx,
		`UPDATE sessions
		 SET revoked_at = NOW(),
		     revoked_reason = 'user_revoked'
		 WHERE id = $1
		   AND user_id = $2
		   AND revoked_at IS NULL
		 RETURNING id`,
		targetSessionID,
		userID,
	).Scan(&revokedSessionID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("revoke session: %w", err)
	}

	event := AuthEvent{
		UserID:    &userID,
		SessionID: &revokedSessionID,
		EventType: AuthEventSessionRevoked,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata: map[string]any{
			"initiating_session_id": currentSessionID,
		},
	}

	if err := InsertAuthEvent(ctx, tx, event); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit session revocation: %w", err)
	}

	return true, nil
}
