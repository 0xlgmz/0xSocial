package monetization

import "testing"

func TestConfigFromEnvDisabledByDefault(t *testing.T) {
	t.Setenv("MONETIZATION_ENABLED", "false")

	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv() error = %v", err)
	}
	if config.Enabled || config.Placements.Feed.Enabled {
		t.Fatalf("ConfigFromEnv() = %+v, want disabled", config)
	}
	if config.Placements.Feed.Format != "fluid" {
		t.Fatalf("feed format = %q, want fluid", config.Placements.Feed.Format)
	}
}

func TestConfigFromEnvEnabled(t *testing.T) {
	t.Setenv("MONETIZATION_ENABLED", "true")
	t.Setenv("ADSENSE_CLIENT_ID", "ca-pub-1234567890123456")
	t.Setenv("ADSENSE_FEED_ENABLED", "true")
	t.Setenv("ADSENSE_FEED_SLOT_ID", "1234567890")
	t.Setenv("ADSENSE_FEED_FORMAT", "fluid")
	t.Setenv("ADSENSE_FEED_LAYOUT_KEY", "-fb+5w+4e-db+86")
	t.Setenv("ADSENSE_FEED_FIRST_AFTER_POSTS", "5")
	t.Setenv("ADSENSE_FEED_REPEAT_EVERY_POSTS", "8")
	t.Setenv("ADSENSE_FEED_MAXIMUM_ADS", "3")

	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv() error = %v", err)
	}
	if !config.Enabled || !config.Placements.Feed.Enabled {
		t.Fatalf("ConfigFromEnv() = %+v, want enabled", config)
	}
	if config.Placements.Feed.FirstAfterPosts != 5 {
		t.Fatalf("firstAfterPosts = %d, want 5", config.Placements.Feed.FirstAfterPosts)
	}
}

func TestConfigFromEnvRejectsIncompleteEnabledConfig(t *testing.T) {
	t.Setenv("MONETIZATION_ENABLED", "true")

	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("ConfigFromEnv() error = nil, want validation error")
	}
}
