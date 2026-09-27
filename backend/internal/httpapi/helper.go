package httpapi

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/mailer"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/media"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/monetization"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool         *pgxpool.Pool
	limiters     ratelimit.AuthLimiters
	events       *postgres.AuthEventRecorder
	mailer       mailer.Sender
	media        media.Store
	monetization monetization.Config
}

func NewHandler(pool *pgxpool.Pool, sender mailer.Sender, mediaStore media.Store, monetizationConfig monetization.Config) *Handler {
	return &Handler{
		pool:         pool,
		limiters:     ratelimit.NewAuthLimiters(),
		events:       postgres.NewAuthEventRecorder(pool),
		mailer:       sender,
		media:        mediaStore,
		monetization: monetizationConfig,
	}
}

func (h *Handler) recordAuthEvent(r *http.Request, event postgres.AuthEvent) error {
	if event.IPAddress == "" {
		event.IPAddress = clientIP(r)
	}

	if event.UserAgent == "" {
		event.UserAgent = r.UserAgent()
	}

	return h.events.Record(r.Context(), event)
}

func (h *Handler) recordAuthEventBestEffort(r *http.Request, event postgres.AuthEvent) {
	if err := h.recordAuthEvent(r, event); err != nil {
		slog.Error(
			"failed to record authentication event",
			"event_type", event.EventType,
			"error", err,
		)
	}
}

func validateHandle(handle string) bool {
	var handlePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,29}$`)
	return handlePattern.MatchString(handle)
}
