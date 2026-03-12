package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidUsername    = errors.New("username must be 3-32 alphanumeric characters")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrLastAdmin          = errors.New("cannot remove the last admin")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrDeleteSelf         = errors.New("cannot delete your own account")
	ErrNotOwner           = errors.New("resource does not belong to user")
)

// AuthService handles authentication and user management
type AuthService struct {
	userRepo         *repository.UserRepo
	refreshTokenRepo *repository.RefreshTokenRepo
	sessionRepo      *repository.SessionRepo
	jwtManager       *auth.JWTManager
	db               *sql.DB
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepo, refreshTokenRepo *repository.RefreshTokenRepo, sessionRepo *repository.SessionRepo, jwtManager *auth.JWTManager, db *sql.DB) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		sessionRepo:      sessionRepo,
		jwtManager:       jwtManager,
		db:               db,
	}
}

// IsConfigured checks if the system has any users (first-run detection)
func (s *AuthService) IsConfigured(ctx context.Context) (bool, error) {
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return false, fmt.Errorf("checking user count: %w", err)
	}
	return count > 0, nil
}

// CreateFirstAdmin creates the first admin user (setup wizard)
func (s *AuthService) CreateFirstAdmin(ctx context.Context, username, password, displayName string) (*model.User, error) {
	// Check if already configured
	configured, err := s.IsConfigured(ctx)
	if err != nil {
		return nil, err
	}
	if configured {
		return nil, errors.New("setup already completed")
	}

	return s.CreateUser(ctx, username, password, displayName, true)
}

// CreateUser creates a new user with validation
func (s *AuthService) CreateUser(ctx context.Context, username, password, displayName string, isAdmin bool) (*model.User, error) {
	// Normalize FIRST
	username = strings.ToLower(strings.TrimSpace(username))

	// Validate AFTER normalize
	if !isValidUsername(username) {
		return nil, ErrInvalidUsername
	}
	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	// Check if username exists
	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("checking username: %w", err)
	}

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Create user
	user := &model.User{
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: hash,
		IsAdmin:      isAdmin,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Grant access to all existing libraries
	if err := s.userRepo.GrantAllLibraries(ctx, user.ID); err != nil {
		log.Printf("warning: granting library access for user %d: %v", user.ID, err)
	}

	return user, nil
}

// Login validates credentials and returns user with tokens
func (s *AuthService) Login(ctx context.Context, username, password, deviceName, ipAddress, userAgent string) (*model.User, *auth.TokenPair, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("fetching user: %w", err)
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.createSession(ctx, user.ID, user.IsAdmin, deviceName, ipAddress, userAgent)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Refresh rotates refresh token and returns new token pair.
// Uses atomic DELETE...RETURNING to prevent concurrent replay attacks.
func (s *AuthService) Refresh(ctx context.Context, refreshToken, deviceName, ipAddress, userAgent string) (*auth.TokenPair, error) {
	tokenHash := auth.HashToken(refreshToken)

	// Atomically consume the token — only one concurrent request can succeed
	rt, err := s.refreshTokenRepo.ConsumeByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("consuming refresh token: %w", err)
	}

	// Clean up associated session
	_ = s.sessionRepo.DeleteByRefreshTokenID(ctx, rt.ID)

	// Check if expired (token already consumed, just reject)
	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	// Create new session
	tokens, err := s.createSession(ctx, user.ID, user.IsAdmin, deviceName, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Logout invalidates a refresh token and deletes associated session
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := auth.HashToken(refreshToken)

	rt, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Already logged out
		}
		return fmt.Errorf("fetching refresh token: %w", err)
	}

	// Delete associated session (by refresh_token_id)
	if err := s.sessionRepo.DeleteByRefreshTokenID(ctx, rt.ID); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	// Delete the refresh token
	if err := s.refreshTokenRepo.Delete(ctx, rt.ID); err != nil {
		return fmt.Errorf("deleting refresh token: %w", err)
	}

	return nil
}

// LogoutAll invalidates all sessions for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID int64) error {
	if err := s.refreshTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("deleting user refresh tokens: %w", err)
	}
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("deleting user sessions: %w", err)
	}
	return nil
}

// createSession creates a new session with tokens
func (s *AuthService) createSession(ctx context.Context, userID int64, isAdmin bool, deviceName, ipAddress, userAgent string) (*auth.TokenPair, error) {
	// Generate refresh token first (opaque, no session ID needed)
	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	// Store refresh token
	tokenHash := auth.HashToken(refreshToken)
	expiresAt := time.Now().Add(auth.RefreshTokenExpiry)

	rtID, err := s.refreshTokenRepo.Create(ctx, userID, tokenHash, deviceName, ipAddress, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	// Create session — now we have the session ID
	sessionExpires := time.Now().Add(auth.RefreshTokenExpiry)
	sessionID, err := s.sessionRepo.Create(ctx, userID, &rtID, deviceName, ipAddress, userAgent, sessionExpires)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	// Generate access token with session ID embedded
	accessToken, err := s.jwtManager.GenerateAccessToken(userID, isAdmin, sessionID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	return &auth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(auth.AccessTokenExpiry),
	}, nil
}

// ListSessions returns all active sessions for a user
func (s *AuthService) ListSessions(ctx context.Context, userID int64) ([]model.Session, error) {
	return s.sessionRepo.ListByUserID(ctx, userID)
}

// RevokeSession revokes a specific session
func (s *AuthService) RevokeSession(ctx context.Context, sessionID, userID int64) error {
	// Verify the session belongs to the user
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("fetching session: %w", err)
	}

	if session.UserID != userID {
		return ErrNotOwner
	}

	// Delete refresh token if exists
	if session.RefreshTokenID != nil {
		_ = s.refreshTokenRepo.Delete(ctx, *session.RefreshTokenID)
	}

	// Delete session
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}

// ChangePassword changes a user's password (requires current password).
// Password update and session invalidation are atomic — both succeed or neither does.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPass, newPass string) error {
	if len(newPass) < 8 {
		return ErrInvalidPassword
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("fetching user: %w", err)
	}

	if !auth.CheckPassword(user.PasswordHash, oldPass) {
		return ErrInvalidCredentials
	}

	newHash, err := auth.HashPassword(newPass)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	// Atomic: update password + invalidate all sessions in one transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.userRepo.WithTx(tx).UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("deleting refresh tokens: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("deleting sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *AuthService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	return user, nil
}

// ListUsers returns all users
func (s *AuthService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.List(ctx)
}

// UpdateUser updates user info (admin only fields like is_admin)
func (s *AuthService) UpdateUser(ctx context.Context, user *model.User) error {
	// Check if user exists
	existing, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("fetching user: %w", err)
	}

	// If changing admin status, check we're not removing the last admin
	if existing.IsAdmin && !user.IsAdmin {
		adminCount, err := s.userRepo.CountAdmins(ctx)
		if err != nil {
			return fmt.Errorf("counting admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	// Preserve password hash (don't update through this method)
	user.PasswordHash = existing.PasswordHash

	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *AuthService) DeleteUser(ctx context.Context, id int64, deletedBy int64) error {
	// Cannot delete self
	if id == deletedBy {
		return ErrDeleteSelf
	}

	// Check if user exists and get admin status
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("fetching user: %w", err)
	}

	// If deleting an admin, check we're not removing the last one
	if user.IsAdmin {
		adminCount, err := s.userRepo.CountAdmins(ctx)
		if err != nil {
			return fmt.Errorf("counting admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	// Delete all sessions first
	_ = s.LogoutAll(ctx, id)

	return s.userRepo.Delete(ctx, id)
}

// SetLibraryAccess sets which libraries a user can access (replaces all existing)
func (s *AuthService) SetLibraryAccess(ctx context.Context, userID int64, libraryIDs []int64) error {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("fetching user: %w", err)
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.userRepo.WithTx(tx).SetLibraryAccess(ctx, userID, libraryIDs); err != nil {
		return fmt.Errorf("setting library access: %w", err)
	}

	return tx.Commit()
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// isValidUsername checks if username is alphanumeric 3-32 chars
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	return usernameRegex.MatchString(username)
}
