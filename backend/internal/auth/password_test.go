package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "securepassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty
		},
		{
			name:     "long password",
			password: "thisisaverylongpasswordthatexceedsnormallengths1234567890",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify the hash is not empty and is different from password
			if hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if hash == tt.password {
				t.Error("HashPassword() returned plaintext")
			}

			// Verify we can check the password
			if !CheckPassword(hash, tt.password) {
				t.Error("CheckPassword() returned false for correct password")
			}

			// Verify wrong password fails
			if CheckPassword(hash, tt.password+"wrong") {
				t.Error("CheckPassword() returned true for wrong password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		plain    string
		expected bool
	}{
		{
			name:     "correct password",
			hash:     hash,
			plain:    "testpassword",
			expected: true,
		},
		{
			name:     "wrong password",
			hash:     hash,
			plain:    "wrongpassword",
			expected: false,
		},
		{
			name:     "empty password",
			hash:     hash,
			plain:    "",
			expected: false,
		},
		{
			name:     "invalid hash",
			hash:     "notahash",
			plain:    "password",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.hash, tt.plain)
			if result != tt.expected {
				t.Errorf("CheckPassword() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
