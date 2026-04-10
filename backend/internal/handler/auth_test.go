package handler

import (
	"testing"

	"github.com/thawng/velox/internal/model"
)

func TestNewUserInfoMapsAvatarToProfilePath(t *testing.T) {
	user := &model.User{
		ID:          42,
		Username:    "velox",
		DisplayName: "Velox User",
		IsAdmin:     true,
		AvatarPath:  "/images/users/42.png",
	}

	got := newUserInfo(user)

	if got.ID != user.ID {
		t.Fatalf("ID = %d, want %d", got.ID, user.ID)
	}
	if got.Username != user.Username {
		t.Fatalf("Username = %q, want %q", got.Username, user.Username)
	}
	if got.DisplayName != user.DisplayName {
		t.Fatalf("DisplayName = %q, want %q", got.DisplayName, user.DisplayName)
	}
	if got.IsAdmin != user.IsAdmin {
		t.Fatalf("IsAdmin = %v, want %v", got.IsAdmin, user.IsAdmin)
	}
	if got.ProfilePath != user.AvatarPath {
		t.Fatalf("ProfilePath = %q, want %q", got.ProfilePath, user.AvatarPath)
	}
}
