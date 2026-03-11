package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/repository"
)

func setupAuthTestDB(t *testing.T) (*sql.DB, *AuthService) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create tables
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

		CREATE TABLE user_library_access (
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			library_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, library_id)
		);

		CREATE TABLE libraries (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE refresh_tokens (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash  TEXT NOT NULL UNIQUE,
			device_name TEXT DEFAULT '',
			ip_address  TEXT DEFAULT '',
			expires_at  DATETIME NOT NULL,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE sessions (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id          INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token_id INTEGER REFERENCES refresh_tokens(id) ON DELETE SET NULL,
			device_name      TEXT DEFAULT '',
			ip_address       TEXT DEFAULT '',
			user_agent       TEXT DEFAULT '',
			expires_at       DATETIME NOT NULL,
			last_active_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		INSERT INTO libraries (id, name) VALUES (1, 'Test Library');
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	userRepo := repository.NewUserRepo(db)
	refreshTokenRepo := repository.NewRefreshTokenRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	jwtManager := auth.NewJWTManager([]byte("test-secret-32-bytes-long-key!!"))
	authSvc := NewAuthService(userRepo, refreshTokenRepo, sessionRepo, jwtManager, db)

	return db, authSvc
}

func TestAuthService_IsConfigured(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Initially not configured
	configured, err := svc.IsConfigured(ctx)
	if err != nil {
		t.Fatalf("IsConfigured() error = %v", err)
	}
	if configured {
		t.Error("IsConfigured() = true, want false")
	}

	// Create user
	_, err = svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin User")
	if err != nil {
		t.Fatalf("CreateFirstAdmin() error = %v", err)
	}

	// Now configured
	configured, err = svc.IsConfigured(ctx)
	if err != nil {
		t.Fatalf("IsConfigured() error = %v", err)
	}
	if !configured {
		t.Error("IsConfigured() = false, want true")
	}
}

func TestAuthService_CreateFirstAdmin_AlreadyConfigured(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create first admin
	_, err := svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")
	if err != nil {
		t.Fatalf("CreateFirstAdmin() error = %v", err)
	}

	// Try to create again
	_, err = svc.CreateFirstAdmin(ctx, "admin2", "password123", "Admin 2")
	if err == nil {
		t.Error("CreateFirstAdmin() should fail when already configured")
	}
}

func TestAuthService_CreateUser_Validation(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()
	// Create first admin to pass configured check
	_, _ = svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")

	tests := []struct {
		name        string
		username    string
		password    string
		displayName string
		wantErr     error
	}{
		{
			name:        "short username",
			username:    "ab",
			password:    "password123",
			displayName: "Test",
			wantErr:     ErrInvalidUsername,
		},
		{
			name:        "short password",
			username:    "testuser",
			password:    "1234567",
			displayName: "Test",
			wantErr:     ErrInvalidPassword,
		},
		{
			name:        "invalid username chars",
			username:    "test@user",
			password:    "password123",
			displayName: "Test",
			wantErr:     ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateUser(ctx, tt.username, tt.password, tt.displayName, false)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_CreateUser_DuplicateUsername(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()
	// Create first admin
	_, _ = svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")

	// Try to create user with same username
	_, err := svc.CreateUser(ctx, "admin", "password456", "Another Admin", false)
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("CreateUser() error = %v, want ErrUserExists", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create user
	_, err := svc.CreateFirstAdmin(ctx, "testuser", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateFirstAdmin() error = %v", err)
	}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{
			name:     "correct credentials",
			username: "testuser",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "wrong password",
			username: "testuser",
			password: "wrongpassword",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name:     "non-existent user",
			username: "nobody",
			password: "password123",
			wantErr:  ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, tokens, err := svc.Login(ctx, tt.username, tt.password, "", "", "")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Login() error = %v, want %v", err, tt.wantErr)
				return
			}
			if err == nil && user.Username != "testuser" {
				t.Errorf("Login() user = %v, want testuser", user.Username)
			}
			if err == nil && tokens == nil {
				t.Error("Login() tokens = nil, want non-nil")
			}
		})
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create user
	user, err := svc.CreateFirstAdmin(ctx, "testuser", "oldpassword123", "Test User")
	if err != nil {
		t.Fatalf("CreateFirstAdmin() error = %v", err)
	}

	// Change password
	err = svc.ChangePassword(ctx, user.ID, "oldpassword123", "newpassword456")
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}

	// Try login with old password
	_, _, err = svc.Login(ctx, "testuser", "oldpassword123", "", "", "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Error("Should not be able to login with old password")
	}

	// Login with new password
	_, _, err = svc.Login(ctx, "testuser", "newpassword456", "", "", "")
	if err != nil {
		t.Errorf("Should be able to login with new password: %v", err)
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	user, err := svc.CreateFirstAdmin(ctx, "testuser", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateFirstAdmin() error = %v", err)
	}

	err = svc.ChangePassword(ctx, user.ID, "wrongpassword", "newpassword456")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ChangePassword() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_DeleteUser(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create two admins
	admin1, _ := svc.CreateFirstAdmin(ctx, "admin1", "password123", "Admin 1")
	admin2, _ := svc.CreateUser(ctx, "admin2", "password123", "Admin 2", true)

	// Delete admin2
	err := svc.DeleteUser(ctx, admin2.ID, admin1.ID)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	// Verify deleted
	_, err = svc.GetUser(ctx, admin2.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Error("User should be deleted")
	}
}

func TestAuthService_DeleteUser_CannotDeleteSelf(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	admin, _ := svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")

	err := svc.DeleteUser(ctx, admin.ID, admin.ID)
	if err == nil {
		t.Error("Should not be able to delete self")
	}
}

func TestAuthService_DeleteUser_LastAdmin(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create admin
	admin, _ := svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")

	// Create non-admin user
	nonAdmin, _ := svc.CreateUser(ctx, "user", "password123", "User", false)

	// Try to delete admin (should fail - last admin)
	err := svc.DeleteUser(ctx, admin.ID, nonAdmin.ID)
	if !errors.Is(err, ErrLastAdmin) {
		t.Errorf("DeleteUser() error = %v, want ErrLastAdmin", err)
	}
}

func TestAuthService_UpdateUser_LastAdmin(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create admin
	admin, _ := svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")

	// Try to demote admin (should fail - last admin)
	admin.IsAdmin = false
	err := svc.UpdateUser(ctx, admin)
	if !errors.Is(err, ErrLastAdmin) {
		t.Errorf("UpdateUser() error = %v, want ErrLastAdmin", err)
	}
}

func TestAuthService_ListUsers(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Initially empty (before first admin)
	users, err := svc.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 0 {
		t.Errorf("ListUsers() = %d users, want 0", len(users))
	}

	// Create users
	_, _ = svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")
	_, _ = svc.CreateUser(ctx, "user1", "password123", "User 1", false)
	_, _ = svc.CreateUser(ctx, "user2", "password123", "User 2", false)

	users, err = svc.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 3 {
		t.Errorf("ListUsers() = %d users, want 3", len(users))
	}
}

func TestAuthService_GetUser(t *testing.T) {
	db, svc := setupAuthTestDB(t)
	defer db.Close()

	ctx := context.Background()

	user, _ := svc.CreateFirstAdmin(ctx, "admin", "password123", "Admin")

	// Get existing
	found, err := svc.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("GetUser() ID = %d, want %d", found.ID, user.ID)
	}

	// Get non-existent
	_, err = svc.GetUser(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetUser() error = %v, want ErrNotFound", err)
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		{"valid", "testuser", true},
		{"min length", "abc", true},
		{"max length", "thisisaverylongusernamethat32chr", true},
		{"too short", "ab", false},
		{"too long", "thisisaverylongusernamethat33chrs", false},
		{"with hyphen", "test-user", true},
		{"with underscore", "test_user", true},
		{"with space", "test user", false},
		{"with at", "test@user", false},
		{"empty", "", false},
		{"unicode", "tëstuser", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidUsername(tt.username); got != tt.valid {
				t.Errorf("isValidUsername(%q) = %v, want %v", tt.username, got, tt.valid)
			}
		})
	}
}
