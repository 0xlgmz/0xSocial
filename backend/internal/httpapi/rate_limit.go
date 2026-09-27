package httpapi

import (
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
)

func clientIP(r *http.Request) string {
	peerHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peerHost = r.RemoteAddr
	}

	peerIP, err := netip.ParseAddr(peerHost)
	if err != nil {
		return peerHost
	}

	peerIP = peerIP.Unmap()

	// X-Real-IP is trusted only because Nginx is the direct peer.
	if peerIP.IsLoopback() {
		forwardedIP, err := netip.ParseAddr(
			strings.TrimSpace(r.Header.Get("X-Real-IP")),
		)
		if err == nil {
			return forwardedIP.Unmap().String()
		}
	}

	// Direct connections never get to choose their identity.
	return peerIP.String()
}

func rateLimited(w http.ResponseWriter, limiter *ratelimit.Limiter, key string) bool {
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
