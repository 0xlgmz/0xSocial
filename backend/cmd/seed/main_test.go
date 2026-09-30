package main

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/media"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/seed"
	"github.com/aws/smithy-go"
)

func TestPostgresURLFromEnvEscapesCredentials(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "db")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "social db")
	t.Setenv("POSTGRES_USER", "seed@user")
	t.Setenv("POSTGRES_PASSWORD", "p@ss/word")
	value := postgresURLFromEnv()
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.Hostname(); got != "db" {
		t.Fatalf("host = %q", got)
	}
	if password, _ := parsed.User.Password(); password != "p@ss/word" {
		t.Fatalf("password was not preserved")
	}
}

type fakeObjectStore struct {
	mu      sync.Mutex
	objects map[string]media.ObjectInfo
}

func (store *fakeObjectStore) PutObjectIfAbsent(_ context.Context, key, contentType string, body []byte) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.objects[key]; exists {
		return false, nil
	}
	store.objects[key] = media.ObjectInfo{ContentType: contentType, SizeBytes: int64(len(body))}
	return true, nil
}

func (store *fakeObjectStore) HeadObject(_ context.Context, key string) (media.ObjectInfo, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.objects[key], nil
}

func (store *fakeObjectStore) DeleteObject(_ context.Context, key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.objects, key)
	return nil
}

func TestUploadMediaCreatesAndReusesObjects(t *testing.T) {
	dataset := seed.Dataset{
		Users: []seed.User{{Ordinal: 0, Topic: 2}},
		Posts: []seed.Post{{UserOrdinal: 0, Content: "image post", CreatedAt: time.Now(), ImageCount: 2}},
	}
	assets := mediaAssets(dataset, "seed", 42)
	store := &fakeObjectStore{objects: map[string]media.ObjectInfo{
		assets[0].ObjectKey: {ContentType: "image/png", SizeBytes: 123},
	}}

	updated, created, err := uploadMedia(context.Background(), store, dataset, assets, 42, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 || created[0] != assets[1].ObjectKey {
		t.Fatalf("created keys = %v", created)
	}
	if updated[0].SizeBytes != 123 || updated[1].SizeBytes <= 0 {
		t.Fatalf("unexpected media sizes: %d, %d", updated[0].SizeBytes, updated[1].SizeBytes)
	}
}

func TestRetryableMediaError(t *testing.T) {
	if !retryableMediaError(&smithy.GenericAPIError{Code: "ServiceUnavailable", Message: "slow down"}) {
		t.Fatal("R2 service throttling should be retried")
	}
	if retryableMediaError(&smithy.GenericAPIError{Code: "AccessDenied", Message: "bad credentials"}) {
		t.Fatal("credential errors should not be retried")
	}
}

func TestLooksLocal(t *testing.T) {
	for _, value := range []string{"postgres://u:p@localhost:5432/db", "postgres://u:p@127.0.0.1/db", "postgres://u:p@db:5432/db"} {
		if !looksLocal(value) {
			t.Errorf("expected local: %s", value)
		}
	}
	if looksLocal("postgres://u:p@database.example.com/db") {
		t.Fatal("remote database identified as local")
	}
}
