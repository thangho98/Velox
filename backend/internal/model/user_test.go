package model

import (
	"encoding/json"
	"testing"
)

func TestUserMarshalJSONIncludesAvatarAndProfilePath(t *testing.T) {
	user := User{
		ID:          7,
		Username:    "tester",
		DisplayName: "Test User",
		IsAdmin:     false,
		// Stored form: local:// scheme (as produced by ImageStorage).
		AvatarPath: "local://user/7/avatar.jpg",
	}

	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// ImagePath.MarshalJSON normalizes local:// to the served /api/images URL.
	const wantNormalized = "/api/images/local/user/7/avatar.jpg"
	if got := payload["avatar_path"]; got != wantNormalized {
		t.Fatalf("avatar_path = %v, want %q", got, wantNormalized)
	}
	if got := payload["profile_path"]; got != wantNormalized {
		t.Fatalf("profile_path = %v, want %q", got, wantNormalized)
	}
	if _, ok := payload["password_hash"]; ok {
		t.Fatal("password_hash should never be serialized")
	}
}
