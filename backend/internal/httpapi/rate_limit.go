package httpapi

import (
	"math"
	"net"
	"net/http"
	"strconv"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
)

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func rateLimited(
	w http.ResponseWriter,
	limiter *ratelimit.Limiter,
	key string,
) bool {
	allowed, retryAfter := limiter.Allow(key)
	if allowed {
		return false
	}

	seconds := int(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}

	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	http.Error(w, "too many requests", http.StatusTooManyRequests)

	return true
}
