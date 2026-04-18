package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"
)

func randKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, KeySize)
	if _, err := rand.Read(k); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return k
}

func TestRoundTrip(t *testing.T) {
	key := randKey(t)
	cases := [][]byte{
		[]byte(""),
		[]byte("hello"),
		[]byte("user@example.com\x00password123"),
		bytes.Repeat([]byte("A"), 4096),
	}
	for i, plaintext := range cases {
		ct, err := Encrypt(key, plaintext)
		if err != nil {
			t.Fatalf("case %d Encrypt: %v", i, err)
		}
		if len(ct) < NonceSize {
			t.Errorf("case %d ciphertext too short: %d", i, len(ct))
		}
		pt, err := Decrypt(key, ct)
		if err != nil {
			t.Fatalf("case %d Decrypt: %v", i, err)
		}
		if !bytes.Equal(pt, plaintext) {
			t.Errorf("case %d mismatch: got %q want %q", i, pt, plaintext)
		}
	}
}

func TestTamperDetection(t *testing.T) {
	key := randKey(t)
	ct, err := Encrypt(key, []byte("secret payload"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	// Flip a bit in the ciphertext body (after the nonce).
	tampered := make([]byte, len(ct))
	copy(tampered, ct)
	tampered[NonceSize+2] ^= 0x80

	_, err = Decrypt(key, tampered)
	if !errors.Is(err, ErrTampered) {
		t.Errorf("err = %v; want ErrTampered", err)
	}
}

func TestWrongKey(t *testing.T) {
	key1 := randKey(t)
	key2 := randKey(t)

	ct, err := Encrypt(key1, []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	_, err = Decrypt(key2, ct)
	if !errors.Is(err, ErrTampered) {
		t.Errorf("err = %v; want ErrTampered", err)
	}
}

func TestKeyValidation(t *testing.T) {
	shortKey := make([]byte, 16) // AES-128 key, but we require 256
	_, err := Encrypt(shortKey, []byte("x"))
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("Encrypt short key err = %v; want ErrInvalidKey", err)
	}
	_, err = Decrypt(shortKey, make([]byte, NonceSize+16))
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("Decrypt short key err = %v; want ErrInvalidKey", err)
	}
}

func TestShortCiphertext(t *testing.T) {
	key := randKey(t)
	_, err := Decrypt(key, []byte("too-short"))
	if !errors.Is(err, ErrShortCiphertext) {
		t.Errorf("err = %v; want ErrShortCiphertext", err)
	}
}

func TestNonceUniqueness(t *testing.T) {
	key := randKey(t)
	ct1, _ := Encrypt(key, []byte("same"))
	ct2, _ := Encrypt(key, []byte("same"))
	if bytes.Equal(ct1, ct2) {
		t.Error("identical ciphertexts imply deterministic nonce — GCM security broken")
	}
}

func TestLoadKeyFromEnv_Success(t *testing.T) {
	// 32 bytes = 64 hex chars
	t.Setenv("VELOX_CLOUD_SECRET", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	key, err := LoadKeyFromEnv()
	if err != nil {
		t.Fatalf("LoadKeyFromEnv: %v", err)
	}
	if len(key) != KeySize {
		t.Errorf("key size = %d; want %d", len(key), KeySize)
	}
}

func TestLoadKeyFromEnv_Missing(t *testing.T) {
	os.Unsetenv("VELOX_CLOUD_SECRET")
	_, err := LoadKeyFromEnv()
	if !errors.Is(err, ErrNoKey) {
		t.Errorf("err = %v; want ErrNoKey", err)
	}
}

func TestLoadKeyFromEnv_InvalidLength(t *testing.T) {
	t.Setenv("VELOX_CLOUD_SECRET", "abc123") // too short
	_, err := LoadKeyFromEnv()
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("err = %v; want ErrInvalidKey", err)
	}
}

func TestLoadKeyFromEnv_InvalidHex(t *testing.T) {
	t.Setenv("VELOX_CLOUD_SECRET", "xxxx")
	_, err := LoadKeyFromEnv()
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("err = %v; want ErrInvalidKey", err)
	}
}

func TestLoadOrGenerateKey_GenerateAndPersist(t *testing.T) {
	os.Unsetenv("TEST_KEY_ENV")
	dir := t.TempDir()
	path := dir + "/subdir/key"

	key1, origin, err := LoadOrGenerateKey("TEST_KEY_ENV", path)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if origin != OriginGenerated {
		t.Errorf("origin = %s; want generated", origin)
	}
	if len(key1) != KeySize {
		t.Errorf("key size = %d; want %d", len(key1), KeySize)
	}

	// Second call should load from file, returning the same key.
	key2, origin, err := LoadOrGenerateKey("TEST_KEY_ENV", path)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if origin != OriginFile {
		t.Errorf("origin = %s; want file", origin)
	}
	if !bytes.Equal(key1, key2) {
		t.Error("second call returned different key — file didn't persist")
	}
}

func TestLoadOrGenerateKey_EnvOverridesFile(t *testing.T) {
	// Env takes precedence even when a valid file exists.
	envKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	t.Setenv("TEST_KEY_ENV", envKey)

	dir := t.TempDir()
	path := dir + "/key"
	// Seed file with a *different* key so we can tell which wins.
	fileKey := "ff" + envKey[2:]
	if err := os.WriteFile(path, []byte(fileKey), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	got, origin, err := LoadOrGenerateKey("TEST_KEY_ENV", path)
	if err != nil {
		t.Fatalf("LoadOrGenerateKey: %v", err)
	}
	if origin != OriginEnv {
		t.Errorf("origin = %s; want env", origin)
	}
	wantKey, _ := hex.DecodeString(envKey)
	if !bytes.Equal(got, wantKey) {
		t.Error("env value should override file")
	}
}

func TestLoadOrGenerateKey_RejectCorruptFile(t *testing.T) {
	os.Unsetenv("TEST_KEY_ENV")
	dir := t.TempDir()
	path := dir + "/key"
	if err := os.WriteFile(path, []byte("not-hex-garbage"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	_, _, err := LoadOrGenerateKey("TEST_KEY_ENV", path)
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("err = %v; want ErrInvalidKey", err)
	}
}

func TestLoadOrGenerateKey_FilePerms(t *testing.T) {
	os.Unsetenv("TEST_KEY_ENV")
	dir := t.TempDir()
	path := dir + "/generated-key"
	if _, _, err := LoadOrGenerateKey("TEST_KEY_ENV", path); err != nil {
		t.Fatalf("generate: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("file perms = %o; want 600", mode)
	}
}
