package ratelimit

import "time"

// IPAndEmailLimiters groups the two keys used to protect an authentication
// endpoint. Keeping them together avoids passing a growing list of limiters
// through the router and handler constructors.
type IPAndEmailLimiters struct {
	IP    *Limiter
	Email *Limiter
}

// AuthLimiters contains the rate-limit policy for authentication endpoints.
type AuthLimiters struct {
	Registration       *Limiter
	Login              IPAndEmailLimiters
	ResendVerification IPAndEmailLimiters
}

func NewAuthLimiters() AuthLimiters {
	return AuthLimiters{
		Registration: New(10, time.Hour),
		Login: IPAndEmailLimiters{
			IP:    New(20, time.Minute),
			Email: New(5, 15*time.Minute),
		},
		ResendVerification: IPAndEmailLimiters{
			IP:    New(5, time.Hour),
			Email: New(3, time.Hour),
		},
	}
}
