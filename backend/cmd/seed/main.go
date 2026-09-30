// Command seed creates a large, deterministic social dataset in an existing
// 0xSocial database. It intentionally uses the same schema as the application
// while keeping all generated accounts in an isolated namespace.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/media"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/seed"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/jackc/pgx/v5"
)

type options struct {
	databaseURL, password, prefix, viewerHandle, confirm string
	users, postsPerUser, maxPosts, followsPerUser, reports, days int
	imagePercent, mediaWorkers                           int
	randomSeed                                           int64
	timeout                                              time.Duration
}

const defaultSeedPassword = "SeedPass123!"

func main() {
	var o options
	flag.StringVar(&o.databaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL URL (or use DATABASE_URL / POSTGRES_* environment variables)")
	flag.StringVar(&o.password, "password", envOr("SEED_PASSWORD", defaultSeedPassword), "password shared by generated accounts")
	flag.StringVar(&o.prefix, "prefix", "seed", "lowercase handle prefix; identifies rows replaced on rerun")
	flag.StringVar(&o.viewerHandle, "viewer-handle", "", "existing account that should follow generated accounts")
	flag.StringVar(&o.confirm, "confirm", "", "required as SEED for a database host other than localhost or db")
	flag.IntVar(&o.users, "users", 300, "number of generated users")
	flag.IntVar(&o.postsPerUser, "posts-per-user", 40, "target posts per user (actual activity varies by 50-150%)")
	flag.IntVar(&o.maxPosts, "max-posts", 0, "hard cap on total posts; 0 leaves the generated total uncapped")
	flag.IntVar(&o.followsPerUser, "follows-per-user", 35, "minimum generated follows per user")
	flag.IntVar(&o.reports, "reports", 120, "moderation reports to generate")
	flag.IntVar(&o.imagePercent, "image-percent", 25, "percentage of active posts with one or more images (0 disables media)")
	flag.IntVar(&o.mediaWorkers, "media-workers", 4, "concurrent image uploads")
	flag.DurationVar(&o.timeout, "timeout", 2*time.Hour, "maximum duration for generation, uploads, and database writes")
	flag.IntVar(&o.days, "days", 365, "period over which account and post activity is distributed")
	flag.Int64Var(&o.randomSeed, "random-seed", 20260929, "random seed for reproducible data")
	flag.Parse()

	config := seed.Config{Users: o.users, PostsPerUser: o.postsPerUser, MaxPosts: o.maxPosts, FollowsPerUser: o.followsPerUser, Reports: o.reports, ImagePercent: o.imagePercent, Days: o.days, Prefix: o.prefix, RandomSeed: o.randomSeed, Now: time.Now()}
	if err := seed.Validate(config); err != nil {
		log.Fatalf("invalid options: %v", err)
	}
	if len(o.password) < 8 {
		log.Fatal("seed password must contain at least 8 characters")
	}
	if o.mediaWorkers < 1 || o.mediaWorkers > 64 {
		log.Fatal("media-workers must be between 1 and 64")
	}
	if o.timeout < time.Minute || o.timeout > 24*time.Hour {
		log.Fatal("timeout must be between 1m and 24h")
	}

	if o.databaseURL == "" {
		o.databaseURL = postgresURLFromEnv()
	}
	if o.databaseURL == "" {
		log.Fatal("database connection is missing; set DATABASE_URL or all POSTGRES_* variables")
	}
	isLocal := looksLocal(o.databaseURL)
	if !isLocal && o.confirm != "SEED" {
		log.Fatal("refusing to modify a non-local database without -confirm SEED")
	}
	if !isLocal && o.password == defaultSeedPassword {
		log.Fatal("refusing to create predictable logins on a non-local database; set a strong SEED_PASSWORD")
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()
	pool, err := postgres.Open(ctx, o.databaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	dataset := seed.Generate(config)
	passwordHash, err := auth.HashPassword(o.password)
	if err != nil {
		log.Fatalf("hash seed password: %v", err)
	}
	assets := mediaAssets(dataset, o.prefix, o.randomSeed)
	var mediaStore *media.R2Store
	newObjectKeys := make([]string, 0)
	if len(assets) > 0 {
		mediaConfig, configured, err := media.R2ConfigFromEnv()
		if err != nil {
			log.Fatalf("invalid media configuration: %v", err)
		}
		if !configured {
			log.Fatal("image seeding requires the R2_* environment variables; use -image-percent 0 to seed text only")
		}
		mediaStore, err = media.NewR2Store(ctx, mediaConfig)
		if err != nil {
			log.Fatalf("configure media storage: %v", err)
		}
		log.Printf("uploading or reusing %d seed images with %d workers", len(assets), o.mediaWorkers)
		assets, newObjectKeys, err = uploadMedia(ctx, mediaStore, dataset, assets, o.randomSeed, o.mediaWorkers)
		if err != nil {
			cleanupMedia(context.Background(), mediaStore, newObjectKeys)
			log.Fatalf("upload seed images: %v", err)
		}
	}
	started := time.Now()
	stats, err := writeDataset(ctx, pool, o, dataset, assets, passwordHash)
	if err != nil {
		cleanupMedia(context.Background(), mediaStore, newObjectKeys)
		log.Fatalf("seed database: %v", err)
	}
	cleanupMedia(context.Background(), mediaStore, stats.staleMediaKeys)

	log.Printf("seed complete in %s: %d users, %d posts, %d images, %d follows, %d reports", time.Since(started).Round(time.Millisecond), stats.users, stats.posts, len(assets), stats.follows, stats.reports)
	if isLocal {
		log.Printf("example login: %s / %s", dataset.Users[0].Email, o.password)
	} else {
		log.Printf("example login email: %s (using the configured seed password)", dataset.Users[0].Email)
	}
	if o.viewerHandle != "" {
		log.Printf("@%s now follows %d generated accounts", o.viewerHandle, stats.viewerFollows)
	}
}

type result struct {
	users, posts, follows, reports, viewerFollows int64
	staleMediaKeys                                []string
}

type imageAsset struct {
	PostOrdinal int
	Position    int
	ObjectKey   string
	SizeBytes   int64
}

type dbPool interface {
	Begin(context.Context) (pgx.Tx, error)
}

func writeDataset(ctx context.Context, pool dbPool, o options, dataset seed.Dataset, assets []imageAsset, passwordHash string) (result, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return result{}, err
	}
	defer tx.Rollback(ctx)

	// Serialize runs using a stable, application-specific advisory lock.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(2026092901)`); err != nil {
		return result{}, err
	}
	oldMediaKeys, err := existingMediaKeys(ctx, tx, o.prefix)
	if err != nil {
		return result{}, err
	}
	if err = removePrevious(ctx, tx, o.prefix); err != nil {
		return result{}, err
	}

	userIDs := make([]int64, len(dataset.Users))
	if _, err = tx.Exec(ctx, `CREATE TEMP TABLE seed_users (
		ordinal INTEGER, email TEXT, created_at TIMESTAMPTZ
	) ON COMMIT DROP`); err != nil {
		return result{}, fmt.Errorf("create seed user staging table: %w", err)
	}
	stagedUsers := make([][]any, 0, len(dataset.Users))
	for _, user := range dataset.Users {
		stagedUsers = append(stagedUsers, []any{user.Ordinal, user.Email, user.CreatedAt})
	}
	if _, err = tx.CopyFrom(ctx, pgx.Identifier{"seed_users"}, []string{"ordinal", "email", "created_at"}, pgx.CopyFromRows(stagedUsers)); err != nil {
		return result{}, fmt.Errorf("stage users: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO users (email, email_normalized, email_verified_at, status, sessions_valid_after, created_at, updated_at)
		SELECT email, email, created_at, 'active', created_at, created_at, created_at FROM seed_users ORDER BY ordinal`); err != nil {
		return result{}, fmt.Errorf("insert users: %w", err)
	}
	rows, err := tx.Query(ctx, `SELECT s.ordinal, u.id FROM seed_users s JOIN users u ON u.email_normalized=s.email ORDER BY s.ordinal`)
	if err != nil {
		return result{}, fmt.Errorf("read inserted users: %w", err)
	}
	for rows.Next() {
		var ordinal int
		var id int64
		if err := rows.Scan(&ordinal, &id); err != nil {
			rows.Close()
			return result{}, err
		}
		userIDs[ordinal] = id
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result{}, err
	}
	rows.Close()

	credentials := make([][]any, 0, len(dataset.Users))
	profiles := make([][]any, 0, len(dataset.Users))
	for _, user := range dataset.Users {
		credentials = append(credentials, []any{userIDs[user.Ordinal], passwordHash, user.CreatedAt})
		profiles = append(profiles, []any{userIDs[user.Ordinal], user.Handle, user.DisplayName, user.Bio, user.AvatarURL, user.CreatedAt, user.CreatedAt})
	}
	if _, err = tx.CopyFrom(ctx, pgx.Identifier{"password_credentials"}, []string{"user_id", "password_hash", "password_changed_at"}, pgx.CopyFromRows(credentials)); err != nil {
		return result{}, fmt.Errorf("insert credentials: %w", err)
	}
	if _, err = tx.CopyFrom(ctx, pgx.Identifier{"profiles"}, []string{"user_id", "handle", "display_name", "bio", "avatar_url", "created_at", "updated_at"}, pgx.CopyFromRows(profiles)); err != nil {
		return result{}, fmt.Errorf("insert profiles: %w", err)
	}

	postRows := make([][]any, 0, len(dataset.Posts))
	for _, post := range dataset.Posts {
		postRows = append(postRows, []any{userIDs[post.UserOrdinal], post.Content, post.CreatedAt, post.CreatedAt, post.DeletedAt})
	}
	postsWritten, err := tx.CopyFrom(ctx, pgx.Identifier{"posts"}, []string{"user_id", "content", "created_at", "updated_at", "deleted_at"}, pgx.CopyFromRows(postRows))
	if err != nil {
		return result{}, fmt.Errorf("insert posts: %w", err)
	}
	postIDs := make([]int64, len(dataset.Posts))
	rows, err = tx.Query(ctx, `SELECT id FROM posts WHERE user_id = ANY($1::BIGINT[]) ORDER BY id`, userIDs)
	if err != nil {
		return result{}, fmt.Errorf("read inserted posts: %w", err)
	}
	postOrdinal := 0
	for rows.Next() {
		if postOrdinal >= len(postIDs) {
			rows.Close()
			return result{}, errors.New("database returned more seeded posts than expected")
		}
		if err := rows.Scan(&postIDs[postOrdinal]); err != nil {
			rows.Close()
			return result{}, err
		}
		postOrdinal++
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result{}, err
	}
	rows.Close()
	if postOrdinal != len(postIDs) {
		return result{}, fmt.Errorf("database returned %d seeded posts, expected %d", postOrdinal, len(postIDs))
	}
	if len(assets) > 0 {
		mediaRows := make([][]any, 0, len(assets))
		for _, asset := range assets {
			post := dataset.Posts[asset.PostOrdinal]
			created := post.CreatedAt
			mediaRows = append(mediaRows, []any{
				userIDs[post.UserOrdinal], postIDs[asset.PostOrdinal], asset.ObjectKey,
				"image/png", asset.SizeBytes, asset.Position, "attached",
				created.AddDate(10, 0, 0), created, created,
			})
		}
		if _, err := tx.CopyFrom(ctx, pgx.Identifier{"post_media"}, []string{"user_id", "post_id", "object_key", "mime_type", "size_bytes", "position", "status", "expires_at", "created_at", "attached_at"}, pgx.CopyFromRows(mediaRows)); err != nil {
			return result{}, fmt.Errorf("insert post media: %w", err)
		}
	}

	followRows := make([][]any, 0, len(dataset.Follows))
	for i, follow := range dataset.Follows {
		created := dataset.Users[follow.FollowerOrdinal].CreatedAt
		if followingCreated := dataset.Users[follow.FollowingOrdinal].CreatedAt; followingCreated.After(created) {
			created = followingCreated
		}
		created = created.Add(time.Duration((i%720)+1) * time.Hour)
		if created.After(time.Now()) {
			created = time.Now()
		}
		followRows = append(followRows, []any{userIDs[follow.FollowerOrdinal], userIDs[follow.FollowingOrdinal], created})
	}
	followsWritten, err := tx.CopyFrom(ctx, pgx.Identifier{"follows"}, []string{"follower_id", "following_id", "created_at"}, pgx.CopyFromRows(followRows))
	if err != nil {
		return result{}, fmt.Errorf("insert follows: %w", err)
	}

	viewerFollows, err := connectViewer(ctx, tx, o.viewerHandle, userIDs)
	if err != nil {
		return result{}, err
	}
	reportsWritten, err := insertReports(ctx, tx, userIDs, o.reports, o.randomSeed)
	if err != nil {
		return result{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return result{}, err
	}
	newKeys := make(map[string]struct{}, len(assets))
	for _, asset := range assets {
		newKeys[asset.ObjectKey] = struct{}{}
	}
	staleKeys := make([]string, 0)
	for _, key := range oldMediaKeys {
		if _, retained := newKeys[key]; !retained {
			staleKeys = append(staleKeys, key)
		}
	}
	return result{users: int64(len(userIDs)), posts: postsWritten, follows: followsWritten, reports: reportsWritten, viewerFollows: viewerFollows, staleMediaKeys: staleKeys}, nil
}

func existingMediaKeys(ctx context.Context, tx pgx.Tx, prefix string) ([]string, error) {
	emailPattern := "seed." + prefix + ".%@seed.0xsocial.local"
	rows, err := tx.Query(ctx, `SELECT pm.object_key FROM post_media pm JOIN users u ON u.id=pm.user_id JOIN profiles p ON p.user_id=u.id WHERE u.email_normalized LIKE $2 AND p.handle LIKE $1`, prefix+"%", emailPattern)
	if err != nil {
		return nil, fmt.Errorf("list old seed media: %w", err)
	}
	defer rows.Close()
	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func removePrevious(ctx context.Context, tx pgx.Tx, prefix string) error {
	// Reports use RESTRICT references, so remove only those touching namespaced users first.
	emailPattern := "seed." + prefix + ".%@seed.0xsocial.local"
	if _, err := tx.Exec(ctx, `DELETE FROM post_reports WHERE reporter_id IN (
		SELECT u.id FROM users u JOIN profiles p ON p.user_id=u.id WHERE u.email_normalized LIKE $2 AND p.handle LIKE $1
	) OR post_id IN (SELECT po.id FROM posts po JOIN users u ON u.id=po.user_id JOIN profiles p ON p.user_id=u.id WHERE u.email_normalized LIKE $2 AND p.handle LIKE $1)`, prefix+"%", emailPattern); err != nil {
		return fmt.Errorf("remove old reports: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM users USING profiles WHERE users.id=profiles.user_id AND users.email_normalized LIKE $2 AND profiles.handle LIKE $1`, prefix+"%", emailPattern); err != nil {
		return fmt.Errorf("remove old seed users: %w", err)
	}
	return nil
}

func connectViewer(ctx context.Context, tx pgx.Tx, handle string, userIDs []int64) (int64, error) {
	if strings.TrimSpace(handle) == "" {
		return 0, nil
	}
	var viewerID int64
	if err := tx.QueryRow(ctx, `SELECT p.user_id FROM profiles p JOIN users u ON u.id=p.user_id WHERE p.handle=$1 AND u.status='active'`, strings.ToLower(strings.TrimSpace(handle))).Scan(&viewerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("viewer @%s is not an active account", handle)
		}
		return 0, err
	}
	rows := make([][]any, 0, len(userIDs))
	for i, id := range userIDs {
		// Follow every account for modest datasets, or a representative 500.
		if len(userIDs) <= 500 || i%max(1, len(userIDs)/500) == 0 {
			rows = append(rows, []any{viewerID, id})
		}
	}
	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE seed_viewer_follows (follower_id BIGINT, following_id BIGINT) ON COMMIT DROP`); err != nil {
		return 0, err
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"seed_viewer_follows"}, []string{"follower_id", "following_id"}, pgx.CopyFromRows(rows)); err != nil {
		return 0, err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO follows (follower_id, following_id) SELECT follower_id, following_id FROM seed_viewer_follows WHERE follower_id <> following_id ON CONFLICT DO NOTHING`)
	if err != nil {
		return 0, fmt.Errorf("connect viewer: %w", err)
	}
	return tag.RowsAffected(), nil
}

func insertReports(ctx context.Context, tx pgx.Tx, userIDs []int64, requested int, randomSeed int64) (int64, error) {
	if requested == 0 {
		return 0, nil
	}
	rows, err := tx.Query(ctx, `SELECT po.id, po.user_id, po.created_at FROM posts po WHERE po.user_id = ANY($1::BIGINT[]) AND po.deleted_at IS NULL ORDER BY po.id`, userIDs)
	if err != nil {
		return 0, err
	}
	type postAuthor struct {
		postID, authorID int64
		createdAt        time.Time
	}
	posts := make([]postAuthor, 0)
	for rows.Next() {
		var p postAuthor
		if err := rows.Scan(&p.postID, &p.authorID, &p.createdAt); err != nil {
			rows.Close()
			return 0, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	if len(posts) == 0 {
		return 0, nil
	}
	reasons := []string{"spam", "harassment", "hate_speech", "violence", "sexual_content", "self_harm", "false_information", "other"}
	rng := rand.New(rand.NewSource(randomSeed + 1))
	seen := map[[2]int64]bool{}
	reportRows := make([][]any, 0, min(requested, len(posts)))
	for len(reportRows) < requested && len(seen) < len(posts)*min(len(userIDs)-1, 4) {
		post := posts[rng.Intn(len(posts))]
		reporter := userIDs[rng.Intn(len(userIDs))]
		if reporter == post.authorID || seen[[2]int64{post.postID, reporter}] {
			continue
		}
		seen[[2]int64{post.postID, reporter}] = true
		now := time.Now().UTC()
		availableHours := max(1, int(now.Sub(post.createdAt).Hours()))
		created := post.createdAt.Add(time.Duration(rng.Intn(availableHours)) * time.Hour)
		reportRows = append(reportRows, []any{post.postID, reporter, reasons[rng.Intn(len(reasons))], created, created})
	}
	return tx.CopyFrom(ctx, pgx.Identifier{"post_reports"}, []string{"post_id", "reporter_id", "reason", "created_at", "updated_at"}, pgx.CopyFromRows(reportRows))
}

func mediaAssets(dataset seed.Dataset, prefix string, randomSeed int64) []imageAsset {
	assets := make([]imageAsset, 0)
	for postOrdinal, post := range dataset.Posts {
		for position := 0; position < post.ImageCount; position++ {
			assets = append(assets, imageAsset{
				PostOrdinal: postOrdinal,
				Position:    position,
				ObjectKey: fmt.Sprintf(
					"posts/seeded/%s/%d/post-%07d-%d.png",
					prefix, randomSeed, postOrdinal, position,
				),
			})
		}
	}
	return assets
}

type seedObjectStore interface {
	PutObjectIfAbsent(context.Context, string, string, []byte) (bool, error)
	HeadObject(context.Context, string) (media.ObjectInfo, error)
	DeleteObject(context.Context, string) error
}

func uploadMedia(ctx context.Context, store seedObjectStore, dataset seed.Dataset, assets []imageAsset, randomSeed int64, workers int) ([]imageAsset, []string, error) {
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	type outcome struct {
		index   int
		created bool
		size    int64
		err     error
	}
	jobs := make(chan int)
	results := make(chan outcome, workers)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				asset := assets[index]
				post := dataset.Posts[asset.PostOrdinal]
				topic := dataset.Users[post.UserOrdinal].Topic
				body, err := seed.RenderImage(asset.PostOrdinal, asset.Position, topic, randomSeed)
				if err != nil {
					results <- outcome{index: index, err: err}
					continue
				}
				var created bool
				err = retryMediaOperation(workerCtx, index, func() error {
					var uploadErr error
					created, uploadErr = store.PutObjectIfAbsent(workerCtx, asset.ObjectKey, "image/png", body)
					return uploadErr
				})
				size := int64(len(body))
				if err == nil && !created {
					var info media.ObjectInfo
					err = retryMediaOperation(workerCtx, index+1, func() error {
						var headErr error
						info, headErr = store.HeadObject(workerCtx, asset.ObjectKey)
						return headErr
					})
					if err == nil && info.ContentType != "image/png" {
						err = fmt.Errorf("existing object %s has content type %s", asset.ObjectKey, info.ContentType)
					}
					if err == nil {
						size = info.SizeBytes
					}
				}
				results <- outcome{index: index, created: created, size: size, err: err}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for index := range assets {
			select {
			case jobs <- index:
			case <-workerCtx.Done():
				return
			}
		}
	}()
	go func() { group.Wait(); close(results) }()

	createdKeys := make([]string, 0, len(assets))
	var firstError error
	for result := range results {
		if result.created {
			createdKeys = append(createdKeys, assets[result.index].ObjectKey)
		}
		if result.err != nil && firstError == nil {
			firstError = result.err
			cancel()
		}
		if result.err == nil {
			assets[result.index].SizeBytes = result.size
		}
	}
	return assets, createdKeys, firstError
}

func cleanupMedia(parent context.Context, store seedObjectStore, objectKeys []string) {
	if store == nil || len(objectKeys) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()
	jobs := make(chan string)
	var group sync.WaitGroup
	for worker := 0; worker < min(2, len(objectKeys)); worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for key := range jobs {
				if err := retryMediaOperation(ctx, len(key), func() error { return store.DeleteObject(ctx, key) }); err != nil {
					log.Printf("warning: could not delete obsolete seed image %s: %v", key, err)
				}
			}
		}()
	}
sendKeys:
	for _, key := range objectKeys {
		select {
		case jobs <- key:
		case <-ctx.Done():
			break sendKeys
		}
	}
	close(jobs)
	group.Wait()
}

func retryMediaOperation(ctx context.Context, jitterSeed int, operation func() error) error {
	const maximumAttempts = 8
	var lastError error
	for attempt := 0; attempt < maximumAttempts; attempt++ {
		if err := operation(); err != nil {
			lastError = err
			if !retryableMediaError(err) || attempt == maximumAttempts-1 {
				return err
			}
			backoff := 500 * time.Millisecond * time.Duration(1<<min(attempt, 4))
			jitter := time.Duration((jitterSeed*97+attempt*193)%400) * time.Millisecond
			timer := time.NewTimer(backoff + jitter)
			select {
			case <-timer.C:
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return ctx.Err()
			}
			continue
		}
		return nil
	}
	return lastError
}

func retryableMediaError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		switch apiError.ErrorCode() {
		case "ServiceUnavailable", "SlowDown", "TooManyRequests", "Throttling", "RequestTimeout", "InternalError":
			return true
		}
	}
	var responseError *smithyhttp.ResponseError
	if errors.As(err, &responseError) {
		status := responseError.HTTPStatusCode()
		return status == 429 || status == 500 || status == 502 || status == 503 || status == 504
	}
	var networkError net.Error
	return errors.As(err, &networkError) && (networkError.Timeout() || networkError.Temporary())
}

func postgresURLFromEnv() string {
	host, db, user, password := os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")
	if host == "" || db == "" || user == "" || password == "" {
		return ""
	}
	port := envOr("POSTGRES_PORT", "5432")
	connection := &url.URL{Scheme: "postgres", User: url.UserPassword(user, password), Host: host + ":" + port, Path: "/" + db}
	query := connection.Query()
	query.Set("sslmode", "disable")
	connection.RawQuery = query.Encode()
	return connection.String()
}

func looksLocal(databaseURL string) bool {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "db" || host == "postgres"
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
