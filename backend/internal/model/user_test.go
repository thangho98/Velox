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
		AvatarPath:  "/avatars/tester.png",
	}

	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got, want := payload["avatar_path"], user.AvatarPath; got != want {
		t.Fatalf("avatar_path = %v, want %q", got, want)
	}
	if got, want := payload["profile_path"], user.AvatarPath; got != want {
		t.Fatalf("profile_path = %v, want %q", got, want)
	}
	if _, ok := payload["password_hash"]; ok {
		t.Fatal("password_hash should never be serialized")
	}
}
