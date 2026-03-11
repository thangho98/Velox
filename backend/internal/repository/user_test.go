package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/model"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create tables for testing
	_, err = db.Exec(`
		CREATE TABLE users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			username      TEXT NOT NULL UNIQUE,
			display_name  TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			is_admin      INTEGER DEFAULT 0,
			avatar_path   TEXT DEFAULT '',
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE user_preferences (
			user_id                INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			subtitle_language      TEXT DEFAULT '',
			audio_language         TEXT DEFAULT '',
			max_streaming_quality  TEXT DEFAULT 'auto',
			theme                  TEXT DEFAULT 'dark'
		);

		CREATE TABLE user_library_access (
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			library_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, library_id)
		);

		CREATE TABLE libraries (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return db
}

func TestUserRepo_Create(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if user.ID == 0 {
		t.Error("Create() did not set ID")
	}
	if user.CreatedAt == "" {
		t.Error("Create() did not set CreatedAt")
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	// Create user
	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get by ID
	found, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if found.Username != user.Username {
		t.Errorf("GetByID() username = %v, want %v", found.Username, user.Username)
	}
	if !found.IsAdmin {
		t.Error("GetByID() IsAdmin = false, want true")
	}
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("GetByID() error = %v, want ErrNoRows", err)
	}
}

func TestUserRepo_GetByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "hashedpassword",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetByUsername() error = %v", err)
	}

	if found.ID != user.ID {
		t.Errorf("GetByUsername() ID = %v, want %v", found.ID, user.ID)
	}
}

func TestUserRepo_Count(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	// Initially 0
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %v, want 0", count)
	}

	// Create user
	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "hashedpassword",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Now 1
	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count() = %v, want 1", count)
	}
}

func TestUserRepo_UpdatePassword(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "oldhash",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.UpdatePassword(ctx, user.ID, "newhash")
	if err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	// Verify
	found, _ := repo.GetByID(ctx, user.ID)
	if found.PasswordHash != "newhash" {
		t.Error("UpdatePassword() did not update hash")
	}
}

func TestUserRepo_Delete(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "hashedpassword",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(ctx, user.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Error("Delete() did not remove user")
	}
}

func TestUserRepo_CountAdmins(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	// Initially 0
	count, err := repo.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("CountAdmins() error = %v", err)
	}
	if count != 0 {
		t.Errorf("CountAdmins() = %v, want 0", count)
	}

	// Create admin
	admin := &model.User{
		Username:     "admin",
		DisplayName:  "Admin",
		PasswordHash: "hash",
		IsAdmin:      true,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	count, err = repo.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("CountAdmins() error = %v", err)
	}
	if count != 1 {
		t.Errorf("CountAdmins() = %v, want 1", count)
	}
}

func TestUserPreferencesRepo_Get(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserPreferencesRepo(db)
	ctx := context.Background()

	// Get non-existent returns defaults
	prefs, err := repo.Get(ctx, 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if prefs.Theme != "dark" {
		t.Errorf("Get() Theme = %v, want dark", prefs.Theme)
	}
	if prefs.MaxStreamingQuality != "auto" {
		t.Errorf("Get() MaxStreamingQuality = %v, want auto", prefs.MaxStreamingQuality)
	}
}

func TestUserPreferencesRepo_Update(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	// Need to create user first
	userRepo := NewUserRepo(db)
	prefsRepo := NewUserPreferencesRepo(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		DisplayName:  "Test",
		PasswordHash: "hash",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	prefs := &model.UserPreferences{
		UserID:              user.ID,
		SubtitleLanguage:    "en",
		AudioLanguage:       "en",
		MaxStreamingQuality: "1080p",
		Theme:               "light",
	}

	err := prefsRepo.Update(ctx, prefs)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	found, err := prefsRepo.Get(ctx, user.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if found.Theme != "light" {
		t.Errorf("Get() Theme = %v, want light", found.Theme)
	}
	if found.SubtitleLanguage != "en" {
		t.Errorf("Get() SubtitleLanguage = %v, want en", found.SubtitleLanguage)
	}
}
