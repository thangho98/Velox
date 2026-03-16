package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
	StreamTokenExpiry  = 2 * time.Hour
)

// Claims represents JWT claims
type Claims struct {
	UserID    int64 `json:"user_id"`
	IsAdmin   bool  `json:"is_admin"`
	SessionID int64 `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTManager handles JWT operations
type JWTManager struct {
	secret []byte
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret []byte) *JWTManager {
	return &JWTManager{secret: secret}
}

// LoadOrCreateSecret loads JWT secret from file or creates a new one
func LoadOrCreateSecret(dataDir string) ([]byte, error) {
	secretPath := filepath.Join(dataDir, ".jwt_secret")

	// Try to load existing secret
	if data, err := os.ReadFile(secretPath); err == nil {
		// Use only first 32 bytes, ignore trailing whitespace/newlines
		if len(data) >= 32 {
			return data[:32], nil
		}
		// File exists but too short - regenerate
	}

	// Generate new secret
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generating JWT secret: %w", err)
	}

	// Save to file
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}
	if err := os.WriteFile(secretPath, secret, 0600); err != nil {
		return nil, fmt.Errorf("saving JWT secret: %w", err)
	}

	return secret, nil
}

// GenerateAccessToken creates a new access token
func (m *JWTManager) GenerateAccessToken(userID int64, isAdmin bool, sessionID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		IsAdmin:   isAdmin,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateStreamToken creates a token for external player URLs (2-hour expiry).
func (m *JWTManager) GenerateStreamToken(userID int64, isAdmin bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(StreamTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefreshToken creates a new refresh token
func (m *JWTManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken hashes a token using SHA256
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateToken validates and parses a JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}

// GenerateTokenPair creates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID int64, isAdmin bool, sessionID int64) (*TokenPair, error) {
	accessToken, err := m.GenerateAccessToken(userID, isAdmin, sessionID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := m.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(AccessTokenExpiry),
	}, nil
}
