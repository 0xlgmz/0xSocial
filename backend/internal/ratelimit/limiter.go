package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count   int
	resetAt time.Time
}

type Limiter struct {
	mu          sync.Mutex
	max         int
	window      time.Duration
	entries     map[string]entry
	lastCleanup time.Time
}

func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		max:         max,
		window:      window,
		entries:     make(map[string]entry),
		lastCleanup: time.Now(),
	}
}

func (l *Limiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastCleanup) >= l.window {
		for key, value := range l.entries {
			if !now.Before(value.resetAt) {
				delete(l.entries, key)
			}
		}

		l.lastCleanup = now
	}

	value, exists := l.entries[key]

	if !exists || !now.Before(value.resetAt) {
		l.entries[key] = entry{
			count:   1,
			resetAt: now.Add(l.window),
		}

		return true, 0
	}

	if value.count >= l.max {
		return false, time.Until(value.resetAt)
	}

	value.count++
	l.entries[key] = value

	return true, 0
}

func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.entries, key)
}
