package monetization

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	clientIDPattern = regexp.MustCompile(`^ca-pub-\d{16}$`)
	slotIDPattern   = regexp.MustCompile(`^\d+$`)
)

type FeedPlacement struct {
	Enabled          bool   `json:"enabled"`
	SlotID           string `json:"slotId"`
	Format           string `json:"format"`
	LayoutKey        string `json:"layoutKey,omitempty"`
	FirstAfterPosts  int    `json:"firstAfterPosts"`
	RepeatEveryPosts int    `json:"repeatEveryPosts"`
	MaximumAds       int    `json:"maximumAds"`
}

type Placements struct {
	Feed FeedPlacement `json:"feed"`
}

type Config struct {
	Enabled    bool       `json:"enabled"`
	ClientID   string     `json:"clientId"`
	Placements Placements `json:"placements"`
}

func DisabledConfig() Config {
	return Config{
		Placements: Placements{
			Feed: FeedPlacement{Format: "fluid"},
		},
	}
}

func ConfigFromEnv() (Config, error) {
	config := DisabledConfig()

	enabled, err := boolEnv("MONETIZATION_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	if !enabled {
		return config, nil
	}

	config.Enabled = true
	config.ClientID = strings.TrimSpace(os.Getenv("ADSENSE_CLIENT_ID"))
	if !clientIDPattern.MatchString(config.ClientID) {
		return Config{}, fmt.Errorf("ADSENSE_CLIENT_ID must match ca-pub- followed by 16 digits")
	}

	feedEnabled, err := boolEnv("ADSENSE_FEED_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	if !feedEnabled {
		return config, nil
	}

	feed := FeedPlacement{
		Enabled:   true,
		SlotID:    strings.TrimSpace(os.Getenv("ADSENSE_FEED_SLOT_ID")),
		Format:    strings.TrimSpace(os.Getenv("ADSENSE_FEED_FORMAT")),
		LayoutKey: strings.TrimSpace(os.Getenv("ADSENSE_FEED_LAYOUT_KEY")),
	}
	if !slotIDPattern.MatchString(feed.SlotID) {
		return Config{}, fmt.Errorf("ADSENSE_FEED_SLOT_ID must contain digits only")
	}
	if feed.Format == "" {
		return Config{}, fmt.Errorf("ADSENSE_FEED_FORMAT is required")
	}

	feed.FirstAfterPosts, err = positiveIntEnv("ADSENSE_FEED_FIRST_AFTER_POSTS")
	if err != nil {
		return Config{}, err
	}
	feed.RepeatEveryPosts, err = positiveIntEnv("ADSENSE_FEED_REPEAT_EVERY_POSTS")
	if err != nil {
		return Config{}, err
	}
	feed.MaximumAds, err = positiveIntEnv("ADSENSE_FEED_MAXIMUM_ADS")
	if err != nil {
		return Config{}, err
	}

	config.Placements.Feed = feed
	return config, nil
}

func boolEnv(name string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", name)
	}
	return value, nil
}

func positiveIntEnv(name string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}
