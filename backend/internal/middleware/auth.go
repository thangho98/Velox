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

// RequireAuth returns a middleware that requires valid JWT authentication.
// If an APIKeyStore is provided, api_key query params are also accepted.
// skipPaths are exact path matches; paths ending with "/*" are treated as prefix matches.
func RequireAuth(jwtManager *auth.JWTManager, apiKeyStore *auth.APIKeyStore, skipPaths ...string) func(http.Handler) http.Handler {
	skipMap := make(map[string]bool)
	var skipPrefixes []string
	for _, path := range skipPaths {
		if strings.HasSuffix(path, "/*") {
			skipPrefixes = append(skipPrefixes, strings.TrimSuffix(path, "*"))
		} else {
			skipMap[path] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped (exact match)
			if skipMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			// Check prefix matches
			for _, prefix := range skipPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Try api_key query param first (short keys for external players)
			if apiKeyStore != nil {
				if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
					if entry := apiKeyStore.Validate(apiKey); entry != nil {
						ctx := auth.ContextWithUser(r.Context(), entry.UserID, entry.IsAdmin)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Extract JWT from Authorization header or token query param
			token := extractToken(r)
			if token == "" {
				respondUnauthorized(w)
				return
			}

			// Validate JWT
			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				respondUnauthorized(w)
				return
			}

			// Add user + session info to context
			ctx := auth.ContextWithSession(r.Context(), claims.UserID, claims.IsAdmin, claims.SessionID)
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

// SessionTracker tracks per-session activity (debounced to max 1 update/minute)
// Call this after RequireAuth to update last_active_at
func SessionTracker(sessionUpdateFunc func(sessionID int64)) func(http.Handler) http.Handler {
	var mu sync.Mutex
	lastUpdates := make(map[int64]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := auth.SessionIDFromContext(r.Context())
			if sessionID > 0 {
				now := time.Now()
				mu.Lock()
				lastUpdate, exists := lastUpdates[sessionID]
				shouldUpdate := !exists || now.Sub(lastUpdate) > time.Minute
				if shouldUpdate {
					lastUpdates[sessionID] = now
				}
				mu.Unlock()

				if shouldUpdate && sessionUpdateFunc != nil {
					go sessionUpdateFunc(sessionID)
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
