package scanner

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/thawng/velox/internal/repository"
)

const verifyBatchSize = 500

// VerifyResult contains verification statistics.
type VerifyResult struct {
	Checked  int
	Missing  int
	Verified int
}

// Verifier checks for missing files and updates the database.
type Verifier struct {
	mediaFileRepo *repository.MediaFileRepo
}

// NewVerifier creates a new verifier.
func NewVerifier(mediaFileRepo *repository.MediaFileRepo) *Verifier {
	return &Verifier{mediaFileRepo: mediaFileRepo}
}

// VerifyLibrary checks all files in a library and marks missing ones.
func (v *Verifier) VerifyLibrary(ctx context.Context, libraryID int64) (*VerifyResult, error) {
	result := &VerifyResult{}
	offset := 0

	for {
		files, err := v.mediaFileRepo.ListByLibraryID(ctx, libraryID, verifyBatchSize, offset)
		if err != nil {
			return result, fmt.Errorf("listing files for library %d: %w", libraryID, err)
		}
		if len(files) == 0 {
			break
		}

		for _, mf := range files {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}

			result.Checked++
			if fileExists(mf.FilePath) {
				if err := v.mediaFileRepo.UpdateLastVerified(ctx, mf.ID); err != nil {
					log.Printf("verify: failed to update last_verified_at for file %d: %v", mf.ID, err)
				}
				result.Verified++
			} else {
				if err := v.mediaFileRepo.MarkMissing(ctx, mf.ID); err != nil {
					log.Printf("verify: failed to mark file %d as missing: %v", mf.ID, err)
				}
				result.Missing++
				log.Printf("verify: missing file %s (id=%d)", mf.FilePath, mf.ID)
			}
		}

		offset += len(files)
		if len(files) < verifyBatchSize {
			break
		}
	}

	return result, nil
}

// VerifyAll checks all files across all libraries.
func (v *Verifier) VerifyAll(ctx context.Context) (*VerifyResult, error) {
	result := &VerifyResult{}
	offset := 0

	log.Println("verify: starting full file verification...")

	for {
		files, err := v.mediaFileRepo.ListAllPaginated(ctx, verifyBatchSize, offset)
		if err != nil {
			return result, fmt.Errorf("listing all files: %w", err)
		}
		if len(files) == 0 {
			break
		}

		for _, mf := range files {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}

			result.Checked++
			if fileExists(mf.FilePath) {
				if err := v.mediaFileRepo.UpdateLastVerified(ctx, mf.ID); err != nil {
					log.Printf("verify: failed to update last_verified_at for file %d: %v", mf.ID, err)
				}
				result.Verified++
			} else {
				if err := v.mediaFileRepo.MarkMissing(ctx, mf.ID); err != nil {
					log.Printf("verify: failed to mark file %d as missing: %v", mf.ID, err)
				}
				result.Missing++
				log.Printf("verify: missing file %s (id=%d)", mf.FilePath, mf.ID)
			}
		}

		offset += len(files)
		if len(files) < verifyBatchSize {
			break
		}
	}

	log.Printf("verify: complete — checked %d, verified %d, missing %d",
		result.Checked, result.Verified, result.Missing)

	return result, nil
}

// VerifyFile checks if a single file exists and updates its status.
func (v *Verifier) VerifyFile(ctx context.Context, fileID int64) (bool, error) {
	file, err := v.mediaFileRepo.GetByID(ctx, fileID)
	if err != nil {
		return false, fmt.Errorf("getting file: %w", err)
	}

	if fileExists(file.FilePath) {
		if err := v.mediaFileRepo.UpdateLastVerified(ctx, fileID); err != nil {
			return true, fmt.Errorf("updating last_verified_at: %w", err)
		}
		return true, nil
	}

	if err := v.mediaFileRepo.MarkMissing(ctx, fileID); err != nil {
		return false, fmt.Errorf("marking missing: %w", err)
	}
	return false, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
