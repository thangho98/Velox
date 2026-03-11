package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/auth"
)

// AuthConfig holds auth middleware configuration
type AuthConfig struct {
	JWTManager *auth.JWTManager
	// SkipPaths are paths that don't require authentication
	SkipPaths map[string]bool
}

// RequireAuth returns a middleware that requires valid JWT authentication
func RequireAuth(jwtManager *auth.JWTManager, skipPaths ...string) func(http.Handler) http.Handler {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			if skipMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header or query param
			token := extractToken(r)
			if token == "" {
				respondUnauthorized(w)
				return
			}

			// Validate token
			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				respondUnauthorized(w)
				return
			}

			// Add user info to context
			ctx := auth.ContextWithUser(r.Context(), claims.UserID, claims.IsAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin returns a middleware that requires admin access
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, isAdmin, ok := auth.UserFromContext(r.Context())
		if !ok || !isAdmin {
			respondForbidden(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth returns a middleware that extracts user info if available but doesn't require it
func OptionalAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token != "" {
				claims, err := jwtManager.ValidateToken(token)
				if err == nil {
					ctx := auth.ContextWithUser(r.Context(), claims.UserID, claims.IsAdmin)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SessionTracker tracks session activity (debounced to max 1 update/minute)
// Call this after RequireAuth to update last_active_at
func SessionTracker(sessionUpdateFunc func(userID int64)) func(http.Handler) http.Handler {
	// Simple debounce: track last update time per user
	var mu sync.Mutex
	lastUpdates := make(map[int64]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _, ok := auth.UserFromContext(r.Context())
			if ok {
				now := time.Now()
				mu.Lock()
				lastUpdate, exists := lastUpdates[userID]
				shouldUpdate := !exists || now.Sub(lastUpdate) > time.Minute
				if shouldUpdate {
					lastUpdates[userID] = now
				}
				mu.Unlock()

				if shouldUpdate && sessionUpdateFunc != nil {
					go sessionUpdateFunc(userID)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	// First try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	// Then try query param (for stream URLs where headers can't be set)
	return r.URL.Query().Get("token")
}

func respondUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}

func respondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"forbidden"}`))
}
