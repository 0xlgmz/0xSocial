package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertRegisteredUser(ctx context.Context, pool *pgxpool.Pool, emailNormalized, hashedPassword string) error {
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
		RETURNING id`,
		emailNormalized,
		emailNormalized,
	).Scan(&userID)

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

	if err := tx.Commit(ctx); err != nil {
		return errors.New("could not create account")
	}
	return nil
}
