package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password using bcrypt with cost 12
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword verifies a plaintext password against a bcrypt hash
func CheckPassword(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
