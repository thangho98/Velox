package auth

import (
	"context"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey int

const (
	userIDKey contextKey = iota
	isAdminKey
	sessionIDKey
)

// UserFromContext extracts user info from context
// Returns (userID, isAdmin, ok) where ok is true if user info exists
func UserFromContext(ctx context.Context) (int64, bool, bool) {
	userID, ok1 := ctx.Value(userIDKey).(int64)
	isAdmin, ok2 := ctx.Value(isAdminKey).(bool)
	return userID, isAdmin, ok1 && ok2
}

// SessionIDFromContext extracts session ID from context (0 if not set)
func SessionIDFromContext(ctx context.Context) int64 {
	sid, _ := ctx.Value(sessionIDKey).(int64)
	return sid
}

// ContextWithUser adds user info to context
func ContextWithUser(ctx context.Context, userID int64, isAdmin bool) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, isAdminKey, isAdmin)
	return ctx
}

// ContextWithSession adds user and session info to context
func ContextWithSession(ctx context.Context, userID int64, isAdmin bool, sessionID int64) context.Context {
	ctx = ContextWithUser(ctx, userID, isAdmin)
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)
	return ctx
}
