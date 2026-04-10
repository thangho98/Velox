package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/thawng/velox/internal/auth"
)

func TestWithOptionalJWTContextAddsUserAndSession(t *testing.T) {
	jwtManager := auth.NewJWTManager([]byte("test-secret-32-bytes-long-key!!"))
	token, err := jwtManager.GenerateAccessToken(99, true, 1234)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	req = withOptionalJWTContext(req, jwtManager)

	userID, isAdmin, ok := auth.UserFromContext(req.Context())
	if !ok {
		t.Fatal("UserFromContext() = not found, want populated user")
	}
	if got, want := userID, int64(99); got != want {
		t.Fatalf("userID = %d, want %d", got, want)
	}
	if !isAdmin {
		t.Fatal("isAdmin = false, want true")
	}
	if got, want := auth.SessionIDFromContext(req.Context()), int64(1234); got != want {
		t.Fatalf("sessionID = %d, want %d", got, want)
	}
}

func TestWithOptionalJWTContextIgnoresInvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager([]byte("test-secret-32-bytes-long-key!!"))
	req := httptest.NewRequest("GET", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	req = withOptionalJWTContext(req, jwtManager)

	if _, _, ok := auth.UserFromContext(req.Context()); ok {
		t.Fatal("UserFromContext() should be empty for invalid token")
	}
	if got := auth.SessionIDFromContext(req.Context()); got != 0 {
		t.Fatalf("sessionID = %d, want 0", got)
	}
}
