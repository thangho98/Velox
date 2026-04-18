package fshare

import "errors"

// Typed errors returned by the fshare client. Callers can match with errors.Is.
var (
	ErrInvalidCredentials = errors.New("fshare: invalid credentials")
	ErrSessionExpired     = errors.New("fshare: session expired (code 201)")
	ErrRateLimit          = errors.New("fshare: rate limited or flaky api")
	ErrFilePassword       = errors.New("fshare: file is password protected")
	ErrLinkDead           = errors.New("fshare: link dead or invalid")
	ErrAppKeyInvalid      = errors.New("fshare: app_key invalid or revoked")
	ErrNotVIP             = errors.New("fshare: account not VIP, download not available")
	ErrCaptchaRequired    = errors.New("fshare: captcha required (login blocked)")
	ErrNotLoggedIn        = errors.New("fshare: client has no active session")
)
