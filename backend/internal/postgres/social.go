package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSocialProfileNotFound = errors.New("social profile not found")
	ErrCannotFollowSelf      = errors.New("cannot follow yourself")
	ErrPostNotFound          = errors.New("post not found")
	ErrCannotReportOwnPost   = errors.New("cannot report your own post")
	ErrDuplicatePostReport   = errors.New("duplicate post report")
)

type Post struct {
	ID          int64
	Handle      string
	DisplayName string
	AvatarURL   *string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	MediaKeys   []string
}

type FollowProfile struct {
	Cursor      int64
	Handle      string
	DisplayName string
	Bio         string
	AvatarURL   *string
	FollowedAt  time.Time
}

type Relationship struct {
	IsSelf     bool
	Following  bool
	FollowedBy bool
}

func CreatePostReport(ctx context.Context, pool *pgxpool.Pool, reporterID, postID int64, reason string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin post report: %w", err)
	}
	defer tx.Rollback(ctx)

	var authorID int64
	err = tx.QueryRow(
		ctx,
		`SELECT user_id
		 FROM posts
		 WHERE id = $1
		   AND deleted_at IS NULL
		 FOR SHARE`,
		postID,
	).Scan(&authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPostNotFound
	}
	if err != nil {
		return fmt.Errorf("find reported post: %w", err)
	}
	if authorID == reporterID {
		return ErrCannotReportOwnPost
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO post_reports (post_id, reporter_id, reason)
		 VALUES ($1, $2, $3)`,
		postID,
		reporterID,
		reason,
	)
	if err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "23505" {
			return ErrDuplicatePostReport
		}
		return fmt.Errorf("create post report: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit post report: %w", err)
	}
	return nil
}

func activeUserIDByHandle(ctx context.Context, pool *pgxpool.Pool, handle string) (int64, error) {
	var userID int64

	err := pool.QueryRow(
		ctx,
		`SELECT p.user_id
		 FROM profiles p
		 JOIN users u ON u.id = p.user_id
		 WHERE p.handle = $1
		   AND u.status = 'active'`,
		handle,
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrSocialProfileNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("find active profile: %w", err)
	}

	return userID, nil
}

func CreatePost(ctx context.Context, pool *pgxpool.Pool, userID int64, content string, media []PendingMedia) (Post, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return Post{}, fmt.Errorf("begin create post: %w", err)
	}
	defer tx.Rollback(ctx)

	var post Post

	err = tx.QueryRow(
		ctx,
		`WITH author AS (
			SELECT user_id, handle, display_name, avatar_url
			FROM profiles
			WHERE user_id = $1
		), inserted AS (
			INSERT INTO posts (user_id, content)
			SELECT user_id, $2
			FROM author
			RETURNING id, user_id, content, created_at, updated_at
		)
		SELECT
			i.id,
			a.handle,
			a.display_name,
			a.avatar_url,
			i.content,
			i.created_at,
			i.updated_at
		FROM inserted i
		JOIN author a ON a.user_id = i.user_id`,
		userID,
		content,
	).Scan(
		&post.ID,
		&post.Handle,
		&post.DisplayName,
		&post.AvatarURL,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return Post{}, fmt.Errorf("create post: %w", err)
	}

	if err := attachPendingMedia(ctx, tx, userID, post.ID, media); err != nil {
		return Post{}, err
	}

	post.MediaKeys = make([]string, 0, len(media))
	for _, pending := range media {
		post.MediaKeys = append(post.MediaKeys, pending.ObjectKey)
	}

	if err := tx.Commit(ctx); err != nil {
		return Post{}, fmt.Errorf("commit create post: %w", err)
	}

	return post, nil
}

func DeletePost(ctx context.Context, pool *pgxpool.Pool, userID, postID int64) ([]string, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin delete post: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(
		ctx,
		`UPDATE posts
		 SET deleted_at = NOW(),
		     updated_at = NOW()
		 WHERE id = $1
		   AND user_id = $2
		   AND deleted_at IS NULL`,
		postID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("delete post: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return nil, ErrPostNotFound
	}

	rows, err := tx.Query(
		ctx,
		`UPDATE post_media
		 SET status = 'deleted'
		 WHERE post_id = $1
		   AND status = 'attached'
		 RETURNING object_key`,
		postID,
	)
	if err != nil {
		return nil, fmt.Errorf("mark post media deleted: %w", err)
	}

	objectKeys := make([]string, 0)
	for rows.Next() {
		var objectKey string
		if err := rows.Scan(&objectKey); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan deleted post media: %w", err)
		}
		objectKeys = append(objectKeys, objectKey)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("read deleted post media: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit delete post: %w", err)
	}

	return objectKeys, nil
}

func ListPostsByHandle(ctx context.Context, pool *pgxpool.Pool, handle string, beforeID int64, limit int) ([]Post, error) {
	userID, err := activeUserIDByHandle(ctx, pool, handle)
	if err != nil {
		return nil, err
	}

	return listPosts(ctx, pool, userID, beforeID, limit, false)
}

func ListFollowingFeed(ctx context.Context, pool *pgxpool.Pool, userID, beforeID int64, limit int) ([]Post, error) {
	return listPosts(ctx, pool, userID, beforeID, limit, true)
}

func ListExplorePosts(ctx context.Context, pool *pgxpool.Pool, beforeID int64, limit int) ([]Post, error) {
	query := `SELECT
			po.id,
			pr.handle,
			pr.display_name,
			pr.avatar_url,
			po.content,
			po.created_at,
			po.updated_at,
			COALESCE(
				ARRAY_AGG(pm.object_key ORDER BY pm.position)
					FILTER (WHERE pm.id IS NOT NULL),
				ARRAY[]::TEXT[]
			)
		 FROM posts po
		 JOIN profiles pr ON pr.user_id = po.user_id
		 JOIN users u ON u.id = po.user_id
		 LEFT JOIN post_media pm
		   ON pm.post_id = po.id
		  AND pm.status = 'attached'
		 WHERE po.deleted_at IS NULL
		   AND u.status = 'active'
		   AND ($1 = 0 OR po.id < $1)
		 GROUP BY po.id, pr.user_id
		 ORDER BY po.id DESC
		 LIMIT $2`

	return queryPosts(ctx, pool, query, limit, beforeID, limit)
}

func listPosts(ctx context.Context, pool *pgxpool.Pool, userID, beforeID int64, limit int, feed bool) ([]Post, error) {
	query := `SELECT
			po.id,
			pr.handle,
			pr.display_name,
			pr.avatar_url,
			po.content,
			po.created_at,
			po.updated_at,
			COALESCE(
				ARRAY_AGG(pm.object_key ORDER BY pm.position)
					FILTER (WHERE pm.id IS NOT NULL),
				ARRAY[]::TEXT[]
			)
		 FROM posts po
		 JOIN profiles pr ON pr.user_id = po.user_id
		 JOIN users u ON u.id = po.user_id
		 LEFT JOIN post_media pm
		   ON pm.post_id = po.id
		  AND pm.status = 'attached'
		 WHERE po.deleted_at IS NULL
		   AND u.status = 'active'
		   AND po.user_id = $1
		   AND ($2 = 0 OR po.id < $2)
		 GROUP BY po.id, pr.user_id
		 ORDER BY po.id DESC
		 LIMIT $3`

	if feed {
		query = `SELECT
				po.id,
				pr.handle,
				pr.display_name,
				pr.avatar_url,
				po.content,
				po.created_at,
				po.updated_at,
				COALESCE(
					ARRAY_AGG(pm.object_key ORDER BY pm.position)
						FILTER (WHERE pm.id IS NOT NULL),
					ARRAY[]::TEXT[]
				)
			 FROM posts po
			 JOIN profiles pr ON pr.user_id = po.user_id
			 JOIN users u ON u.id = po.user_id
			 LEFT JOIN post_media pm
			   ON pm.post_id = po.id
			  AND pm.status = 'attached'
			 WHERE po.deleted_at IS NULL
			   AND u.status = 'active'
			   AND (
				po.user_id = $1
				OR EXISTS (
					SELECT 1
					FROM follows f
					WHERE f.follower_id = $1
					  AND f.following_id = po.user_id
				)
			   )
			   AND ($2 = 0 OR po.id < $2)
			 GROUP BY po.id, pr.user_id
			 ORDER BY po.id DESC
			 LIMIT $3`
	}

	return queryPosts(ctx, pool, query, limit, userID, beforeID, limit)
}

func queryPosts(ctx context.Context, pool *pgxpool.Pool, query string, limit int, args ...any) ([]Post, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]Post, 0, limit)
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.Handle,
			&post.DisplayName,
			&post.AvatarURL,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.MediaKeys,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}

		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read posts: %w", err)
	}

	return posts, nil
}

func FollowUser(ctx context.Context, pool *pgxpool.Pool, followerID int64, handle string) error {
	followingID, err := activeUserIDByHandle(ctx, pool, handle)
	if err != nil {
		return err
	}
	if followerID == followingID {
		return ErrCannotFollowSelf
	}

	_, err = pool.Exec(
		ctx,
		`INSERT INTO follows (follower_id, following_id)
		 VALUES ($1, $2)
		 ON CONFLICT (follower_id, following_id) DO NOTHING`,
		followerID,
		followingID,
	)
	if err != nil {
		return fmt.Errorf("follow user: %w", err)
	}

	return nil
}

func UnfollowUser(ctx context.Context, pool *pgxpool.Pool, followerID int64, handle string) error {
	followingID, err := activeUserIDByHandle(ctx, pool, handle)
	if err != nil {
		return err
	}
	if followerID == followingID {
		return ErrCannotFollowSelf
	}

	_, err = pool.Exec(
		ctx,
		`DELETE FROM follows
		 WHERE follower_id = $1
		   AND following_id = $2`,
		followerID,
		followingID,
	)
	if err != nil {
		return fmt.Errorf("unfollow user: %w", err)
	}

	return nil
}

func GetRelationship(ctx context.Context, pool *pgxpool.Pool, userID int64, handle string) (Relationship, error) {
	targetID, err := activeUserIDByHandle(ctx, pool, handle)
	if err != nil {
		return Relationship{}, err
	}

	relationship := Relationship{IsSelf: userID == targetID}
	if relationship.IsSelf {
		return relationship, nil
	}

	err = pool.QueryRow(
		ctx,
		`SELECT
			EXISTS (
				SELECT 1 FROM follows
				WHERE follower_id = $1 AND following_id = $2
			),
			EXISTS (
				SELECT 1 FROM follows
				WHERE follower_id = $2 AND following_id = $1
			)`,
		userID,
		targetID,
	).Scan(&relationship.Following, &relationship.FollowedBy)
	if err != nil {
		return Relationship{}, fmt.Errorf("get relationship: %w", err)
	}

	return relationship, nil
}

func ListFollowers(ctx context.Context, pool *pgxpool.Pool, handle string, beforeID int64, limit int) ([]FollowProfile, error) {
	userID, err := activeUserIDByHandle(ctx, pool, handle)
	if err != nil {
		return nil, err
	}

	return listFollowProfiles(ctx, pool, userID, beforeID, limit, true)
}

func ListFollowing(ctx context.Context, pool *pgxpool.Pool, handle string, beforeID int64, limit int) ([]FollowProfile, error) {
	userID, err := activeUserIDByHandle(ctx, pool, handle)
	if err != nil {
		return nil, err
	}

	return listFollowProfiles(ctx, pool, userID, beforeID, limit, false)
}

func listFollowProfiles(ctx context.Context, pool *pgxpool.Pool, userID, beforeID int64, limit int, followers bool) ([]FollowProfile, error) {
	query := `SELECT
			f.id,
			p.handle,
			p.display_name,
			p.bio,
			p.avatar_url,
			f.created_at
		 FROM follows f
		 JOIN profiles p ON p.user_id = f.follower_id
		 JOIN users u ON u.id = p.user_id
		 WHERE f.following_id = $1
		   AND u.status = 'active'
		   AND ($2 = 0 OR f.id < $2)
		 ORDER BY f.id DESC
		 LIMIT $3`

	if !followers {
		query = `SELECT
				f.id,
				p.handle,
				p.display_name,
				p.bio,
				p.avatar_url,
				f.created_at
			 FROM follows f
			 JOIN profiles p ON p.user_id = f.following_id
			 JOIN users u ON u.id = p.user_id
			 WHERE f.follower_id = $1
			   AND u.status = 'active'
			   AND ($2 = 0 OR f.id < $2)
			 ORDER BY f.id DESC
			 LIMIT $3`
	}

	rows, err := pool.Query(ctx, query, userID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("list follow profiles: %w", err)
	}
	defer rows.Close()

	profiles := make([]FollowProfile, 0, limit)
	for rows.Next() {
		var profile FollowProfile
		if err := rows.Scan(
			&profile.Cursor,
			&profile.Handle,
			&profile.DisplayName,
			&profile.Bio,
			&profile.AvatarURL,
			&profile.FollowedAt,
		); err != nil {
			return nil, fmt.Errorf("scan follow profile: %w", err)
		}

		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read follow profiles: %w", err)
	}

	return profiles, nil
}
