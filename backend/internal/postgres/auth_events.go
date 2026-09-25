package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	AuthEventRegistrationSucceeded  = "registration_succeeded"
	AuthEventEmailVerified          = "email_verified"
	AuthEventLoginSucceeded         = "login_succeeded"
	AuthEventLoginFailed            = "login_failed"
	AuthEventLogout                 = "logout"
	AuthEventPasswordResetRequested = "password_reset_requested"
	AuthEventPasswordResetSucceeded = "password_reset_succeeded"
	AuthEventSessionRevoked         = "session_revoked"
)

type AuthEvent struct {
	UserID    *int64
	SessionID *int64
	EventType string
	IPAddress string
	UserAgent string
	Metadata  map[string]any
}
type AuthEventExecutor interface {
	Exec(
		context.Context,
		string,
		...any,
	) (pgconn.CommandTag, error)
}
type AuthEventRecorder struct {
	pool *pgxpool.Pool
}

func NewAuthEventRecorder(pool *pgxpool.Pool) *AuthEventRecorder {
	return &AuthEventRecorder{
		pool: pool,
	}
}

func InsertAuthEvent(ctx context.Context, db AuthEventExecutor, event AuthEvent) error {
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}

	_, err := db.Exec(
		ctx,
		`INSERT INTO auth_events (
			user_id,
			session_id,
			event_type,
			ip_address,
			user_agent,
			metadata
		)
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, '')::INET,
			$5,
			$6
		)`,
		event.UserID,
		event.SessionID,
		event.EventType,
		event.IPAddress,
		event.UserAgent,
		event.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert authentication event: %w", err)
	}

	return nil
}

func (recorder *AuthEventRecorder) Record(ctx context.Context, event AuthEvent) error {
	return InsertAuthEvent(ctx, recorder.pool, event)
}

func (recorder *AuthEventRecorder) RecordWith(ctx context.Context, db AuthEventExecutor, event AuthEvent) error {
	return InsertAuthEvent(ctx, db, event)
}
