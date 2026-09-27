package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
	Handle         string
	DisplayName    string
	Bio            string
	AvatarURL      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PostCount      int64
	FollowerCount  int64
	FollowingCount int64
}
type UserPublicProfile struct {
	Handle         string
	DisplayName    string
	Bio            string
	AvatarURL      *string
	PostCount      int64
	FollowerCount  int64
	FollowingCount int64
}

var ErrPublicProfileNotFound = errors.New("public profile not found")

func GetUserProfile(ctx context.Context, pool *pgxpool.Pool, userID int64) (UserProfile, error) {
	var profile UserProfile

	err := pool.QueryRow(
		ctx,
		`SELECT handle, display_name, bio, avatar_url, created_at, updated_at,
		        (SELECT COUNT(*) FROM posts WHERE user_id = profiles.user_id AND deleted_at IS NULL),
		        (SELECT COUNT(*)
		         FROM follows f JOIN users u ON u.id = f.follower_id
		         WHERE f.following_id = profiles.user_id AND u.status = 'active'),
		        (SELECT COUNT(*)
		         FROM follows f JOIN users u ON u.id = f.following_id
		         WHERE f.follower_id = profiles.user_id AND u.status = 'active')
		 FROM profiles
		 WHERE user_id = $1`,
		userID,
	).Scan(
		&profile.Handle,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.PostCount,
		&profile.FollowerCount,
		&profile.FollowingCount,
	)
	if err != nil {
		return UserProfile{}, fmt.Errorf("get user profile: %w", err)
	}

	return profile, nil
}

func UpdateUserProfile(ctx context.Context, pool *pgxpool.Pool, userID int64, displayName, bio *string) (UserProfile, error) {
	var profile UserProfile

	err := pool.QueryRow(
		ctx,
		`WITH updated AS (
			UPDATE profiles
			SET display_name = COALESCE($2::TEXT, display_name),
			    bio = COALESCE($3::TEXT, bio),
			    updated_at = NOW()
			WHERE user_id = $1
			RETURNING user_id, handle, display_name, bio, avatar_url, created_at, updated_at
		)
		SELECT
			u.handle,
			u.display_name,
			u.bio,
			u.avatar_url,
			u.created_at,
			u.updated_at,
			(SELECT COUNT(*) FROM posts WHERE user_id = u.user_id AND deleted_at IS NULL),
			(SELECT COUNT(*)
			 FROM follows f JOIN users account ON account.id = f.follower_id
			 WHERE f.following_id = u.user_id AND account.status = 'active'),
			(SELECT COUNT(*)
			 FROM follows f JOIN users account ON account.id = f.following_id
			 WHERE f.follower_id = u.user_id AND account.status = 'active')
		FROM updated u`,
		userID,
		displayName,
		bio,
	).Scan(
		&profile.Handle,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.PostCount,
		&profile.FollowerCount,
		&profile.FollowingCount,
	)
	if err != nil {
		return UserProfile{}, fmt.Errorf("update user profile: %w", err)
	}

	return profile, nil
}

func GetUserPublicProfile(ctx context.Context, pool *pgxpool.Pool, userHandle string) (UserPublicProfile, error) {
	var profile UserPublicProfile

	err := pool.QueryRow(
		ctx,
		`SELECT
			p.handle,
			p.display_name,
			p.bio,
			p.avatar_url
			,(SELECT COUNT(*) FROM posts WHERE user_id = p.user_id AND deleted_at IS NULL)
			,(SELECT COUNT(*)
			  FROM follows f JOIN users account ON account.id = f.follower_id
			  WHERE f.following_id = p.user_id AND account.status = 'active')
			,(SELECT COUNT(*)
			  FROM follows f JOIN users account ON account.id = f.following_id
			  WHERE f.follower_id = p.user_id AND account.status = 'active')
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		WHERE p.handle = $1
		AND u.status = 'active'`,
		userHandle,
	).Scan(
		&profile.Handle,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.PostCount,
		&profile.FollowerCount,
		&profile.FollowingCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserPublicProfile{}, ErrPublicProfileNotFound
	}
	if err != nil {
		return UserPublicProfile{}, fmt.Errorf("get public profile: %w", err)
	}

	return profile, nil
}
