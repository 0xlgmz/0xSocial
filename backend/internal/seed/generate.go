package seed

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

var topics = []struct {
	Name  string
	Bios  []string
	Posts []string
}{
	{"technology", []string{"Building useful things, one commit at a time.", "Software, systems, and the people behind them.", "Curious about AI, infrastructure, and product craft."}, []string{"Shipped a small improvement today: %s. The boring details really do compound.", "Question for builders: how are you thinking about %s this year?", "A useful lesson from this week: %s works best when the feedback loop is short.", "Hot take: the best tool for %s is usually the one your team can understand tomorrow."}},
	{"design", []string{"Designer of calm, useful interfaces.", "Collecting typography, color, and product details.", "Making complex things feel a little simpler."}, []string{"Today’s design study: %s. Tiny spacing changes made the whole screen breathe.", "I keep coming back to %s as an example of restraint done well.", "A good interface should explain itself. %s is my current test case.", "Sketching ideas around %s — clarity first, decoration second."}},
	{"science", []string{"Always asking one more question.", "Notes from the intersection of data and discovery.", "Science enthusiast; evidence over vibes."}, []string{"Reading about %s and the scale of it is difficult to comprehend.", "Today’s rabbit hole: %s. What result surprised you most?", "A reminder from %s: uncertainty is information, not failure.", "The most interesting part of %s is the question it opens next."}},
	{"outdoors", []string{"Fresh air, long trails, early starts.", "Usually outside. Occasionally posting about it.", "Hikes, bikes, weather, and good coffee."}, []string{"Morning report: %s, clear skies, and nobody else on the trail.", "Planning a weekend around %s. Send your route recommendations.", "A slow afternoon with %s was exactly what I needed.", "Field note: always pack one more layer for %s."}},
	{"food", []string{"Cooking, tasting, and taking notes.", "In search of an excellent lunch.", "Recipes, restaurants, and strong opinions about bread."}, []string{"Made %s tonight and finally got the texture right.", "The secret to %s turned out to be patience and more acidity.", "What is your non-negotiable ingredient for %s?", "Still thinking about the %s I tried this weekend."}},
	{"books", []string{"Reader, annotator, library wanderer.", "Books, essays, and unfinished thoughts.", "Reading widely and recommending selectively."}, []string{"A line of thought from %s has stayed with me all day.", "Just finished %s. The middle was quiet, but the ending earned it.", "What should I read after %s? Looking for something unexpected.", "Revisiting %s and noticing an entirely different book this time."}},
	{"music", []string{"Albums, live rooms, and late-night playlists.", "Listening closely.", "Finding the song that changes the shape of the day."}, []string{"The production on %s sounds incredible with headphones.", "Today’s repeat track is %s. No skips so far.", "There is nothing quite like hearing %s played live.", "Building a playlist around %s — recommendations welcome."}},
	{"community", []string{"Local stories and better neighborhoods.", "Here to listen, learn, and help.", "Community organizer and enthusiastic neighbor."}, []string{"Great turnout for %s today. Thanks to everyone who made time.", "We are looking for a few volunteers for %s next week.", "Small win: %s is funded for another season.", "What would make %s more accessible to everyone?"}},
}

var firstNames = []string{"Avery", "Maya", "Leo", "Nora", "Theo", "Zoe", "Milo", "Iris", "Aria", "Ezra", "Lina", "Noah", "Sofia", "Kai", "Mina", "Owen", "Ada", "Ravi", "June", "Sam", "Elena", "Mateo", "Amara", "Finn"}
var lastNames = []string{"Reed", "Chen", "Patel", "Silva", "Morgan", "Kim", "Okafor", "Rossi", "Khan", "Costa", "Nguyen", "Martin", "Lopez", "Ito", "Brown", "Singh", "Garcia", "Wilson"}
var subjects = []string{"better defaults", "local-first software", "urban gardens", "fermentation", "open data", "trail maps", "public libraries", "small teams", "creative routines", "night trains", "deep work", "seasonal cooking", "community radio", "accessible design", "climate models", "independent publishing"}

type Config struct {
	Users          int
	PostsPerUser   int
	MaxPosts       int
	FollowsPerUser int
	Reports        int
	ImagePercent   int
	Days           int
	Prefix         string
	RandomSeed     int64
	Now            time.Time
}

type User struct {
	Ordinal     int
	Email       string
	Handle      string
	DisplayName string
	Bio         string
	AvatarURL   string
	Topic       int
	CreatedAt   time.Time
}

type Post struct {
	UserOrdinal int
	Content     string
	CreatedAt   time.Time
	DeletedAt   *time.Time
	ImageCount  int
}

type Follow struct{ FollowerOrdinal, FollowingOrdinal int }

type Dataset struct {
	Users   []User
	Posts   []Post
	Follows []Follow
}

func Generate(config Config) Dataset {
	rng := rand.New(rand.NewSource(config.RandomSeed))
	now := config.Now.UTC()
	users := make([]User, 0, config.Users)
	for i := 0; i < config.Users; i++ {
		topic := i % len(topics)
		first := firstNames[(i+rng.Intn(len(firstNames)))%len(firstNames)]
		last := lastNames[(i*7+rng.Intn(len(lastNames)))%len(lastNames)]
		created := now.Add(-time.Duration(24*(30+rng.Intn(max(1, config.Days-29)))) * time.Hour)
		users = append(users, User{
			Ordinal: i, Email: fmt.Sprintf("seed.%s.%06d@seed.0xsocial.local", config.Prefix, i+1),
			Handle: fmt.Sprintf("%s%06d", config.Prefix, i+1), DisplayName: first + " " + last,
			Bio:       topics[topic].Bios[rng.Intn(len(topics[topic].Bios))],
			AvatarURL: fmt.Sprintf("https://api.dicebear.com/9.x/notionists/svg?seed=%s-%d", config.Prefix, i+1),
			Topic:     topic, CreatedAt: created,
		})
	}

	posts := make([]Post, 0, config.Users*config.PostsPerUser)
	for _, user := range users {
		// Uneven activity gives ranking experiments meaningful heavy and light authors.
		count := config.PostsPerUser/2 + rng.Intn(config.PostsPerUser+1)
		for j := 0; j < count; j++ {
			ageHours := rng.Intn(max(24, config.Days*24))
			created := now.Add(-time.Duration(ageHours)*time.Hour - time.Duration(rng.Intn(3600))*time.Second)
			if created.Before(user.CreatedAt) {
				created = user.CreatedAt.Add(time.Duration(rng.Intn(72)+1) * time.Hour)
			}
			template := topics[user.Topic].Posts[rng.Intn(len(topics[user.Topic].Posts))]
			content := fmt.Sprintf(template, subjects[rng.Intn(len(subjects))])
			if rng.Intn(5) == 0 {
				content += fmt.Sprintf(" #%s", topics[user.Topic].Name)
			}
			var deletedAt *time.Time
			if rng.Intn(100) < 2 {
				deleted := created.Add(time.Duration(rng.Intn(72)+1) * time.Hour)
				if deleted.After(now) {
					deleted = now
				}
				deletedAt = &deleted
			}
			imageCount := 0
			if deletedAt == nil && rng.Intn(100) < config.ImagePercent {
				imageCount = 1
				roll := rng.Intn(100)
				if roll < 12 {
					imageCount = 2
				} else if roll < 15 {
					imageCount = 3 + rng.Intn(2)
				}
			}
			posts = append(posts, Post{UserOrdinal: user.Ordinal, Content: content, CreatedAt: created, DeletedAt: deletedAt, ImageCount: imageCount})
		}
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].CreatedAt.Before(posts[j].CreatedAt) })
	if config.MaxPosts > 0 && len(posts) > config.MaxPosts {
		capped := make([]Post, config.MaxPosts)
		if config.MaxPosts == 1 {
			capped[0] = posts[len(posts)-1]
		} else {
			// Sample the full chronological range instead of keeping only the
			// newest posts, which would distort recency-based feed experiments.
			for i := range capped {
				index := i * (len(posts) - 1) / (config.MaxPosts - 1)
				capped[i] = posts[index]
			}
		}
		posts = capped
	}

	pairs := make(map[[2]int]struct{}, config.Users*config.FollowsPerUser)
	followCounts := make([]int, config.Users)
	add := func(a, b int) {
		if a == b {
			return
		}
		key := [2]int{a, b}
		if _, exists := pairs[key]; !exists {
			pairs[key] = struct{}{}
			followCounts[a]++
		}
	}
	hubCount := min(12, max(1, config.Users/25))
	for follower := 0; follower < config.Users; follower++ {
		for followCounts[follower] < min(config.FollowsPerUser, config.Users-1) {
			var following int
			switch n := rng.Intn(100); {
			case n < 20:
				following = rng.Intn(hubCount)
			case n < 75:
				following = (follower % len(topics)) + len(topics)*rng.Intn((config.Users+len(topics)-1)/len(topics))
				if following >= config.Users {
					continue
				}
			default:
				following = rng.Intn(config.Users)
			}
			add(follower, following)
		}
	}
	basePairs := make([][2]int, 0, len(pairs))
	for pair := range pairs {
		basePairs = append(basePairs, pair)
	}
	sort.Slice(basePairs, func(i, j int) bool {
		if basePairs[i][0] == basePairs[j][0] {
			return basePairs[i][1] < basePairs[j][1]
		}
		return basePairs[i][0] < basePairs[j][0]
	})
	for _, pair := range basePairs {
		if rng.Intn(100) < 18 {
			add(pair[1], pair[0])
		}
	}
	follows := make([]Follow, 0, len(pairs))
	for pair := range pairs {
		follows = append(follows, Follow{pair[0], pair[1]})
	}
	sort.Slice(follows, func(i, j int) bool {
		if follows[i].FollowerOrdinal == follows[j].FollowerOrdinal {
			return follows[i].FollowingOrdinal < follows[j].FollowingOrdinal
		}
		return follows[i].FollowerOrdinal < follows[j].FollowerOrdinal
	})
	return Dataset{Users: users, Posts: posts, Follows: follows}
}

func Validate(config Config) error {
	if config.Users < 2 || config.Users > 100_000 {
		return fmt.Errorf("users must be between 2 and 100000")
	}
	if config.PostsPerUser < 1 || config.PostsPerUser > 1_000 {
		return fmt.Errorf("posts-per-user must be between 1 and 1000")
	}
	if int64(config.Users)*int64(config.PostsPerUser) > 5_000_000 {
		return fmt.Errorf("users * posts-per-user cannot exceed 5000000")
	}
	if config.MaxPosts < 0 || config.MaxPosts > 5_000_000 {
		return fmt.Errorf("max-posts must be between 0 and 5000000")
	}
	if config.FollowsPerUser < 0 || config.FollowsPerUser >= config.Users {
		return fmt.Errorf("follows-per-user must be between 0 and users-1")
	}
	if int64(config.Users)*int64(config.FollowsPerUser) > 10_000_000 {
		return fmt.Errorf("users * follows-per-user cannot exceed 10000000")
	}
	if config.Reports < 0 {
		return fmt.Errorf("reports cannot be negative")
	}
	if config.Reports > 1_000_000 {
		return fmt.Errorf("reports cannot exceed 1000000")
	}
	if config.ImagePercent < 0 || config.ImagePercent > 100 {
		return fmt.Errorf("image-percent must be between 0 and 100")
	}
	if config.Days < 30 || config.Days > 3_650 {
		return fmt.Errorf("days must be between 30 and 3650")
	}
	if len(config.Prefix) < 3 || len(config.Prefix) > 15 {
		return fmt.Errorf("prefix must contain 3 to 15 lowercase letters")
	}
	for _, r := range config.Prefix {
		if r < 'a' || r > 'z' {
			return fmt.Errorf("prefix must contain only lowercase letters")
		}
	}
	if strings.TrimSpace(config.Prefix) != config.Prefix {
		return fmt.Errorf("prefix cannot contain whitespace")
	}
	return nil
}
