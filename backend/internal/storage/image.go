package storage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// ImageStorage manages local image files under {dataDir}/images/.
type ImageStorage struct {
	dataDir string
}

// NewImageStorage creates a new ImageStorage rooted at the given data directory.
func NewImageStorage(dataDir string) *ImageStorage {
	return &ImageStorage{dataDir: dataDir}
}

// Save processes and saves an image. Returns the local:// path for DB storage.
// entityType is "media" or "series". imageType is "poster" or "backdrop".
func (s *ImageStorage) Save(entityType string, id int64, imageType string, data []byte) (string, error) {
	// Validate MIME type from magic bytes
	mime := http.DetectContentType(data)
	switch mime {
	case "image/jpeg", "image/png", "image/webp":
		// OK
	default:
		return "", fmt.Errorf("unsupported image format: %s", mime)
	}

	// Decode
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decoding image: %w", err)
	}

	// Resize based on image type
	var maxW, maxH int
	switch imageType {
	case "poster":
		maxW, maxH = 1000, 1500
	case "backdrop":
		maxW, maxH = 1920, 1080
	default:
		return "", fmt.Errorf("unsupported image type: %s", imageType)
	}

	bounds := img.Bounds()
	if bounds.Dx() > maxW || bounds.Dy() > maxH {
		img = imaging.Fit(img, maxW, maxH, imaging.Lanczos)
	}

	// Ensure directory exists
	dir := s.dir(entityType, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating image dir: %w", err)
	}

	filename := imageType + ".jpg"
	absPath := filepath.Join(dir, filename)

	// Write as JPEG quality 90
	f, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("creating image file: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		return "", fmt.Errorf("encoding jpeg: %w", err)
	}

	localPath := fmt.Sprintf("local://%d/%s", id, filename)
	return localPath, nil
}

// Delete removes a local image file.
func (s *ImageStorage) Delete(entityType string, id int64, imageType string) error {
	filename := imageType + ".jpg"
	absPath := filepath.Join(s.dir(entityType, id), filename)
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}

// AbsPath returns the absolute filesystem path for a local image.
func (s *ImageStorage) AbsPath(entityType string, id int64, filename string) string {
	return filepath.Join(s.dir(entityType, id), filename)
}

// Exists checks if a local image exists.
func (s *ImageStorage) Exists(entityType string, id int64, imageType string) bool {
	filename := imageType + ".jpg"
	absPath := filepath.Join(s.dir(entityType, id), filename)
	_, err := os.Stat(absPath)
	return err == nil
}

// Open opens a local image for reading. Caller must close the reader.
func (s *ImageStorage) Open(entityType string, id int64, filename string) (io.ReadCloser, error) {
	absPath := filepath.Join(s.dir(entityType, id), filename)
	return os.Open(absPath)
}

func (s *ImageStorage) dir(entityType string, id int64) string {
	return filepath.Join(s.dataDir, "images", entityType, fmt.Sprintf("%d", id))
}
