package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMediaNotAttachable = errors.New("media is not available to attach")

type PendingMedia struct {
	ID        int64
	ObjectKey string
	MimeType  string
	SizeBytes int64
	ExpiresAt time.Time
}

func CreatePendingMedia(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID int64,
	objectKey string,
	mimeType string,
	sizeBytes int64,
) (PendingMedia, error) {
	var pending PendingMedia

	err := pool.QueryRow(
		ctx,
		`INSERT INTO post_media (
			user_id,
			object_key,
			mime_type,
			size_bytes,
			expires_at
		)
		VALUES ($1, $2, $3, $4, NOW() + INTERVAL '15 minutes')
		RETURNING id, object_key, mime_type, size_bytes, expires_at`,
		userID,
		objectKey,
		mimeType,
		sizeBytes,
	).Scan(
		&pending.ID,
		&pending.ObjectKey,
		&pending.MimeType,
		&pending.SizeBytes,
		&pending.ExpiresAt,
	)
	if err != nil {
		return PendingMedia{}, fmt.Errorf("create pending media: %w", err)
	}

	return pending, nil
}

func DeletePendingMedia(ctx context.Context, pool *pgxpool.Pool, userID, mediaID int64) error {
	_, err := pool.Exec(
		ctx,
		`DELETE FROM post_media
		 WHERE id = $1
		   AND user_id = $2
		   AND status = 'pending'`,
		mediaID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete pending media: %w", err)
	}

	return nil
}

func GetPendingMedia(ctx context.Context, pool *pgxpool.Pool, userID int64, mediaIDs []int64) ([]PendingMedia, error) {
	if len(mediaIDs) == 0 {
		return []PendingMedia{}, nil
	}

	rows, err := pool.Query(
		ctx,
		`SELECT id, object_key, mime_type, size_bytes, expires_at
		 FROM post_media
		 WHERE user_id = $1
		   AND id = ANY($2::BIGINT[])
		   AND status = 'pending'
		   AND expires_at > NOW()`,
		userID,
		mediaIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("get pending media: %w", err)
	}
	defer rows.Close()

	byID := make(map[int64]PendingMedia, len(mediaIDs))
	for rows.Next() {
		var pending PendingMedia
		if err := rows.Scan(
			&pending.ID,
			&pending.ObjectKey,
			&pending.MimeType,
			&pending.SizeBytes,
			&pending.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("scan pending media: %w", err)
		}

		byID[pending.ID] = pending
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read pending media: %w", err)
	}

	ordered := make([]PendingMedia, 0, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		pending, exists := byID[mediaID]
		if !exists {
			return nil, ErrMediaNotAttachable
		}
		ordered = append(ordered, pending)
	}

	return ordered, nil
}

func attachPendingMedia(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	postID int64,
	media []PendingMedia,
) error {
	for position, pending := range media {
		tag, err := tx.Exec(
			ctx,
			`UPDATE post_media
			 SET post_id = $1,
			     position = $2,
			     status = 'attached',
			     attached_at = NOW()
			 WHERE id = $3
			   AND user_id = $4
			   AND status = 'pending'
			   AND expires_at > NOW()`,
			postID,
			position,
			pending.ID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("attach media: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return ErrMediaNotAttachable
		}
	}

	return nil
}
