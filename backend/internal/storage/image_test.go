package storage_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/storage"
)

// minimalValidPNG returns a valid 1x1 transparent PNG
func minimalValidPNG() []byte {
	// This is a valid 1x1 transparent PNG encoded in base64
	data := must(base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="))
	return data
}

func must(buf []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return buf
}

func TestNewImageStorage(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/data")
	assert.NotNil(t, s)
}

func TestImageStorage_AbsPath(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/data")
	path := s.AbsPath("media", 123, "poster.jpg")
	assert.Equal(t, filepath.Join("/tmp/data", "images", "media", "123", "poster.jpg"), path)
}

func TestImageStorage_Dir(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/data")
	// Access dir through AbsPath since dir is unexported
	path := s.AbsPath("series", 456, "backdrop.jpg")
	assert.Contains(t, path, "series")
	assert.Contains(t, path, "456")
	assert.Contains(t, path, "backdrop.jpg")
}

func TestImageMetaResult(t *testing.T) {
	t.Parallel()
	meta := storage.ImageMetaResult{
		Blurhash: "LEHV6nWB2yk8pyo0adR*.7kCMdnj",
		Width:    1920,
		Height:   1080,
	}
	assert.Equal(t, "LEHV6nWB2yk8pyo0adR*.7kCMdnj", meta.Blurhash)
	assert.Equal(t, 1920, meta.Width)
	assert.Equal(t, 1080, meta.Height)
}

func TestImageStorage_Save_UnsupportedFormat(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/test-storage")

	// Text data - unsupported format
	_, _, err := s.Save("media", 1, "poster", []byte("not an image"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported image format")
}

func TestImageStorage_Save_UnsupportedImageType(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/test-storage")

	// Use valid PNG data but with an unsupported image type
	validPNG := minimalValidPNG()
	_, _, err := s.Save("media", 1, "unknown", validPNG)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported image type")
}

func TestImageStorage_Delete_NonExistent(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/test-storage-nonexistent")

	// Delete should not error if file doesn't exist
	err := s.Delete("media", 999, "poster")
	// os.Remove returns nil if file doesn't exist with os.IsNotExist check in Delete
	assert.NoError(t, err)
}

func TestImageStorage_Exists_NonExistent(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/test-storage-nonexistent")

	exists := s.Exists("media", 999, "poster")
	assert.False(t, exists)
}

func TestImageStorage_Open_NonExistent(t *testing.T) {
	t.Parallel()
	s := storage.NewImageStorage("/tmp/test-storage-nonexistent")

	_, err := s.Open("media", 999, "poster.jpg")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err), "should be os.IsNotExist")
}
