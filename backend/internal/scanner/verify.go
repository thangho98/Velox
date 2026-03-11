package scanner

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/thawng/velox/internal/repository"
)

// Verifier checks for missing files and updates database
type Verifier struct {
	mediaFileRepo *repository.MediaFileRepo
}

// NewVerifier creates a new verifier
func NewVerifier(mediaFileRepo *repository.MediaFileRepo) *Verifier {
	return &Verifier{mediaFileRepo: mediaFileRepo}
}

// VerifyLibrary checks all files in a library and marks missing ones
func (v *Verifier) VerifyLibrary(ctx context.Context, libraryID int64) (*VerifyResult, error) {
	// Get all media files for this library
	// Note: This requires joining with media table to filter by library_id
	// For now, we'll need to get all files and filter
	// In production, add a method to repository: ListByLibraryID

	result := &VerifyResult{
		Checked: 0,
		Missing: 0,
	}

	// TODO: Implement full verification once we have ListByLibraryID in MediaFileRepo
	// For now, this is a placeholder

	return result, nil
}

// VerifyResult contains verification statistics
type VerifyResult struct {
	Checked int
	Missing int
	Fixed   int
}

// VerifyFile checks if a single file exists and updates its status
func (v *Verifier) VerifyFile(ctx context.Context, fileID int64) (bool, error) {
	file, err := v.mediaFileRepo.GetByID(ctx, fileID)
	if err != nil {
		return false, fmt.Errorf("getting file: %w", err)
	}

	exists := fileExists(file.FilePath)

	if !exists {
		// File is missing - mark it
		if err := v.mediaFileRepo.MarkMissing(ctx, fileID); err != nil {
			return false, fmt.Errorf("marking missing: %w", err)
		}
		return false, nil
	}

	return exists, nil
}

// FindMissing returns all files that no longer exist on disk
func (v *Verifier) FindMissing(ctx context.Context, limit int) ([]int64, error) {
	// This requires scanning all files - might be slow for large libraries
	// Consider doing this in batches

	var missing []int64

	// TODO: Implement once we have pagination in List methods

	return missing, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// RunFullVerification performs a complete verification of all files
func RunFullVerification(ctx context.Context, mediaFileRepo *repository.MediaFileRepo) (*VerifyResult, error) {
	_ = NewVerifier(mediaFileRepo) // Will be used when implementing full verification
	result := &VerifyResult{}

	log.Println("Starting full file verification...")

	// TODO: Iterate through all files in batches
	// For each file:
	//   1. Check if path exists
	//   2. If not, mark as missing
	//   3. If yes, update last_verified_at

	log.Printf("Verification complete: checked %d, missing %d, fixed %d",
		result.Checked, result.Missing, result.Fixed)

	return result, nil
}
