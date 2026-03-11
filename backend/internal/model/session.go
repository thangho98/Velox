package model

import "time"

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TokenHash  string    `json:"-"`
	DeviceName string    `json:"device_name"`
	IPAddress  string    `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// Session represents a session in the database
type Session struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	RefreshTokenID *int64    `json:"refresh_token_id,omitempty"`
	DeviceName     string    `json:"device_name"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastActiveAt   time.Time `json:"last_active_at"`
	CreatedAt      time.Time `json:"created_at"`
}
