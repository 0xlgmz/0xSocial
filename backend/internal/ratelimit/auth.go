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
	ForgotPassword     IPAndEmailLimiters
	DeleteSession      *Limiter
	VerifyEmail        *Limiter
	PasswordReset      *Limiter
	GetProfile         *Limiter
	GetPublicProfile   *Limiter
	UpdateProfile      *Limiter
	CreatePost         *Limiter
	DeletePost         *Limiter
	ReportPost         *Limiter
	FollowMutation     *Limiter
	SocialRead         *Limiter
	MediaUpload        *Limiter
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
		ForgotPassword: IPAndEmailLimiters{
			IP:    New(5, time.Hour),
			Email: New(3, time.Hour),
		},
		DeleteSession:    New(20, time.Hour),
		VerifyEmail:      New(10, 20*time.Minute),
		PasswordReset:    New(10, 30*time.Minute),
		GetProfile:       New(20, time.Minute),
		GetPublicProfile: New(20, time.Minute),
		UpdateProfile:    New(20, time.Hour),
		CreatePost:       New(30, time.Minute),
		DeletePost:       New(30, time.Minute),
		ReportPost:       New(20, time.Hour),
		FollowMutation:   New(60, time.Minute),
		SocialRead:       New(120, time.Minute),
		MediaUpload:      New(30, time.Hour),
	}
}
