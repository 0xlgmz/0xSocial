# Seeding 0xSocial

`cmd/seed` fills an existing, migrated 0xSocial PostgreSQL database with a
deterministic social graph suitable for feed and ranking work. The defaults
create roughly 12,000 posts across 300 active accounts, at least 10,500 follow
edges, varied profile personas, soft-deleted posts, and 120 moderation reports.
By default, 25% of active posts receive one or more real PNG images.

The data is intentionally non-uniform: users belong to topic communities,
some accounts become network hubs, some follows are reciprocal, author activity
varies, and posts are spread over the previous year. Post insertion order tracks
creation time so the platform's current ID cursor remains meaningful.

## Local database

Image seeding uses the same R2/S3-compatible storage configured for the API.
Export `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`,
and `R2_PUBLIC_BASE_URL` (plus `R2_ENDPOINT` when using a custom endpoint).
The bucket's public URL must be readable by the frontend.

From `backend/`, export the same database variables used by the API and run:

```sh
go run ./cmd/seed -viewer-handle your_handle
```

If media storage is intentionally disabled, retain the older text-only behavior
with `-image-percent 0`.

Alternatively, provide a connection URL directly:

```sh
DATABASE_URL='postgres://postgres:password@localhost:5432/0xsocial?sslmode=disable' \
  go run ./cmd/seed -viewer-handle your_handle
```

`-viewer-handle` is optional. When present, that existing active account follows
the generated accounts, which makes its home feed useful immediately. Without
it, the explore feed and all generated accounts still contain the full dataset.

Every generated account can log in. The default password is `SeedPass123!` and
the first default email is `seed.seed.000001@seed.0xsocial.local`. Set a different
shared password with `SEED_PASSWORD` or `-password`.

## Larger datasets

For about 100,000 posts and a denser graph:

```sh
go run ./cmd/seed \
  -users 1000 \
  -posts-per-user 100 \
  -follows-per-user 75 \
  -reports 1000 \
  -image-percent 30 \
  -days 730 \
  -viewer-handle your_handle
```

Run `go run ./cmd/seed -help` for all sizing and reproducibility flags.

## Reruns and safety

The command runs in one transaction and is safe to rerun. It removes and
recreates only generated accounts matching both the seed email domain and the
selected handle prefix. Existing accounts and their content are left intact.
Use a different `-prefix` to keep multiple independent datasets.

Image objects use deterministic keys. Reruns reuse matching objects, remove
obsolete images after a successful database commit, and remove newly uploaded
objects if the database transaction fails. `-media-workers` controls upload
concurrency and defaults to 4 to avoid object-store throttling. Transient R2
rate-limit and service errors are retried with bounded exponential backoff.
Some image posts contain two to four images so gallery layouts and ranking
features can be exercised as well. Large media runs have a two-hour default
deadline, configurable with `-timeout`.

For a database host other than `localhost`, `127.0.0.1`, `db`, or `postgres`,
the command refuses to run unless the explicit guard is supplied:

```sh
DATABASE_URL='postgres://...' go run ./cmd/seed -confirm SEED
```

A non-local run also requires a non-default `SEED_PASSWORD`; prefer the
environment variable so the password is not exposed in shell process listings
or history. The password is not printed for non-local databases.

Generated avatars use deterministic DiceBear URLs. Seeded post images are
locally rendered deterministic 640x360 PNGs and uploaded to the configured
bucket; the seeder does not depend on a third-party stock-photo service.
