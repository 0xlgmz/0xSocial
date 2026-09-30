package seed

import (
	"reflect"
	"testing"
	"time"
)

func TestGenerateIsDeterministicAndValid(t *testing.T) {
	config := Config{Users: 40, PostsPerUser: 8, FollowsPerUser: 6, Days: 90, Prefix: "seed", RandomSeed: 42, Now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	a, b := Generate(config), Generate(config)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("Generate returned different data for the same seed")
	}
	if len(a.Users) != config.Users {
		t.Fatalf("got %d users", len(a.Users))
	}
	if len(a.Posts) < config.Users*config.PostsPerUser/2 {
		t.Fatalf("too few posts: %d", len(a.Posts))
	}
	seen := map[[2]int]bool{}
	for _, follow := range a.Follows {
		if follow.FollowerOrdinal == follow.FollowingOrdinal {
			t.Fatal("generated self-follow")
		}
		key := [2]int{follow.FollowerOrdinal, follow.FollowingOrdinal}
		if seen[key] {
			t.Fatal("generated duplicate follow")
		}
		seen[key] = true
	}
}

func TestValidate(t *testing.T) {
	valid := Config{Users: 10, PostsPerUser: 2, FollowsPerUser: 3, Days: 30, Prefix: "seed"}
	if err := Validate(valid); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	valid.Prefix = "Seed-"
	if err := Validate(valid); err == nil {
		t.Fatal("invalid prefix accepted")
	}
}

func TestGenerateIncludesTextAndImagePosts(t *testing.T) {
	config := Config{Users: 20, PostsPerUser: 20, FollowsPerUser: 4, ImagePercent: 35, Days: 90, Prefix: "seed", RandomSeed: 99, Now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	dataset := Generate(config)
	var textPosts, imagePosts, galleries int
	for _, post := range dataset.Posts {
		switch {
		case post.ImageCount == 0:
			textPosts++
		case post.ImageCount == 1:
			imagePosts++
		default:
			imagePosts++
			galleries++
		}
	}
	if textPosts == 0 || imagePosts == 0 || galleries == 0 {
		t.Fatalf("expected a mixed dataset, got text=%d image=%d galleries=%d", textPosts, imagePosts, galleries)
	}
}
