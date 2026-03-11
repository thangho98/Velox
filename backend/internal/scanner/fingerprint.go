package scanner

import (
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/cespare/xxhash/v2"
)

const fingerprintHeaderSize = 64 * 1024 // 64KB

// ComputeFingerprint generates a fingerprint for a file.
// Format: "{file_size}:{xxhash64_of_first_64KB}"
// This is path-independent and survives rename/move operations.
func ComputeFingerprint(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stating file: %w", err)
	}

	fileSize := info.Size()

	// Open file
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	// Compute xxhash64 of first 64KB
	h := xxhash.New()
	buf := make([]byte, fingerprintHeaderSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("reading file header: %w", err)
	}

	if n > 0 {
		h.Write(buf[:n])
	}

	return fmt.Sprintf("%d:%x", fileSize, h.Sum64()), nil
}

// ComputeFingerprintWithHasher computes fingerprint using an existing hasher
// Useful for batch operations to reuse hash state
func ComputeFingerprintWithHasher(path string, h hash.Hash64) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stating file: %w", err)
	}

	fileSize := info.Size()

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	h.Reset()
	buf := make([]byte, fingerprintHeaderSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("reading file header: %w", err)
	}

	if n > 0 {
		h.Write(buf[:n])
	}

	return fmt.Sprintf("%d:%x", fileSize, h.Sum64()), nil
}

// ParseFingerprint extracts size and hash from a fingerprint string
func ParseFingerprint(fp string) (size int64, hash string, err error) {
	var hashValue uint64
	_, err = fmt.Sscanf(fp, "%d:%x", &size, &hashValue)
	if err != nil {
		return 0, "", fmt.Errorf("parsing fingerprint: %w", err)
	}
	return size, fmt.Sprintf("%x", hashValue), nil
}
