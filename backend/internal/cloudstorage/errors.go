package cloudstorage

import "errors"

// Errors returned by providers and drivers. Consumers match with errors.Is.
// Driver implementations translate provider-native errors (e.g. fshare
// ErrSessionExpired) into these shared errors at the package boundary.
var (
	ErrUnknownProvider    = errors.New("cloudstorage: unknown provider type")
	ErrDuplicateDriver    = errors.New("cloudstorage: driver type already registered")
	ErrSessionExpired     = errors.New("cloudstorage: session expired")
	ErrInvalidCredentials = errors.New("cloudstorage: invalid credentials")
	ErrRateLimit          = errors.New("cloudstorage: rate limited")
	ErrFileNotFound       = errors.New("cloudstorage: file not found")
	ErrInvalidURL         = errors.New("cloudstorage: url not recognized for this provider")
	ErrNotSupported       = errors.New("cloudstorage: operation not supported")
	ErrAccountNotPremium  = errors.New("cloudstorage: account tier does not support downloads")
)
