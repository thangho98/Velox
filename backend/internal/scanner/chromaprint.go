package scanner

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"os/exec"
	"strconv"
	"strings"
)

// SamplesPerSecond is the approximate number of chromaprint samples per second of audio.
// Each uint32 fingerprint point covers ~0.1238 seconds.
const SamplesPerSecond = 1.0 / 0.1238

// MatchSegment represents a matching audio region between two files.
type MatchSegment struct {
	StartA float64 // start in file A (seconds, relative to region start)
	EndA   float64 // end in file A
	StartB float64 // start in file B
	EndB   float64 // end in file B
	Score  float64 // 0.0-1.0 quality score (1.0 = perfect match)
}

// CheckFpcalc returns true if fpcalc binary is available on the system.
func CheckFpcalc() bool {
	_, err := exec.LookPath("fpcalc")
	return err == nil
}

// ExtractFingerprint runs fpcalc on a file and returns raw fingerprint data.
// startSec and durationSec define the region to analyze.
func ExtractFingerprint(ctx context.Context, filePath string, startSec, durationSec float64) ([]uint32, error) {
	args := []string{
		"-raw",
		"-length", strconv.FormatFloat(durationSec, 'f', 0, 64),
	}
	if startSec > 0 {
		args = append(args, "-offset", strconv.FormatFloat(startSec, 'f', 0, 64))
	}
	args = append(args, filePath)

	cmd := exec.CommandContext(ctx, "fpcalc", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fpcalc failed: %w", err)
	}

	return parseFpcalcOutput(string(out))
}

// parseFpcalcOutput parses fpcalc -raw output into uint32 array.
// Format: DURATION=N\nFINGERPRINT=N,N,N,...
func parseFpcalcOutput(output string) ([]uint32, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "FINGERPRINT=") {
			continue
		}
		raw := strings.TrimPrefix(line, "FINGERPRINT=")
		if raw == "" {
			return nil, nil
		}
		parts := strings.Split(raw, ",")
		fp := make([]uint32, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			// fpcalc outputs signed int32 values, parse as int64 then convert
			val, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing fingerprint value %q: %w", p, err)
			}
			fp = append(fp, uint32(val))
		}
		return fp, nil
	}
	return nil, fmt.Errorf("no FINGERPRINT line in fpcalc output")
}

// ExtractIntroRegion extracts fingerprint for the intro search region.
// Region: min(25% of duration, 600 seconds) from start.
func ExtractIntroRegion(ctx context.Context, filePath string, totalDuration float64) ([]uint32, float64, error) {
	regionDur := math.Min(totalDuration*0.25, 600)
	if regionDur < 30 {
		return nil, 0, nil // Too short to analyze
	}
	fp, err := ExtractFingerprint(ctx, filePath, 0, regionDur)
	return fp, regionDur, err
}

// ExtractCreditsRegion extracts fingerprint for the credits search region.
// Region: last 300 seconds (5 minutes).
func ExtractCreditsRegion(ctx context.Context, filePath string, totalDuration float64) ([]uint32, float64, error) {
	regionDur := math.Min(totalDuration, 300)
	startSec := math.Max(0, totalDuration-300)
	if regionDur < 30 {
		return nil, 0, nil
	}
	fp, err := ExtractFingerprint(ctx, filePath, startSec, regionDur)
	return fp, regionDur, err
}

// CompareFingerprints finds matching audio segments between two fingerprints.
// maxBitDiff is the maximum Hamming distance between two samples to consider a match (typically 10-12).
// Returns the best matching segment, or nil if no match found.
func CompareFingerprints(a, b []uint32, maxBitDiff int) *MatchSegment {
	if len(a) < 20 || len(b) < 20 {
		return nil // Too short for meaningful comparison
	}

	// Step 1: Build shift histogram
	// For efficiency, sample every Nth point in a, check all of b
	step := 1
	if len(a) > 500 {
		step = len(a) / 500 // Sample ~500 points from a
	}

	shiftCounts := make(map[int]int)
	for i := 0; i < len(a); i += step {
		for j := 0; j < len(b); j++ {
			if bits.OnesCount32(a[i]^b[j]) <= maxBitDiff {
				shift := j - i
				shiftCounts[shift]++
			}
		}
	}

	if len(shiftCounts) == 0 {
		return nil
	}

	// Step 2: Find best shift (most matching points)
	bestShift := 0
	bestCount := 0
	for shift, count := range shiftCounts {
		if count > bestCount {
			bestCount = count
			bestShift = shift
		}
	}

	// Need minimum number of matching points to consider valid
	minMatches := 10
	if bestCount < minMatches {
		return nil
	}

	// Step 3: Find longest contiguous matching region at best shift
	return findContiguousMatch(a, b, bestShift, maxBitDiff)
}

// findContiguousMatch finds the longest contiguous matching region at a given shift.
func findContiguousMatch(a, b []uint32, shift, maxBitDiff int) *MatchSegment {
	// Determine the overlap range
	startA := 0
	if shift < 0 {
		startA = -shift
	}
	startB := 0
	if shift > 0 {
		startB = shift
	}

	overlapLen := min(len(a)-startA, len(b)-startB)
	if overlapLen < 20 {
		return nil
	}

	// Walk the overlap, track contiguous matching runs
	var bestRunStart, bestRunLen int
	currentRunStart := 0
	currentRunLen := 0
	gapTolerance := 3 // Allow up to 3 non-matching samples in a row

	for i := 0; i < overlapLen; i++ {
		idxA := startA + i
		idxB := startB + i
		diff := bits.OnesCount32(a[idxA] ^ b[idxB])

		if diff <= maxBitDiff {
			if currentRunLen == 0 {
				currentRunStart = i
			}
			currentRunLen++
		} else {
			// Check if this is just a brief gap
			gapLen := 0
			for g := 1; g <= gapTolerance && i+g < overlapLen; g++ {
				if bits.OnesCount32(a[startA+i+g]^b[startB+i+g]) <= maxBitDiff {
					gapLen = g
					break
				}
			}
			if gapLen > 0 && currentRunLen > 0 {
				currentRunLen += gapLen + 1
				i += gapLen
			} else {
				if currentRunLen > bestRunLen {
					bestRunLen = currentRunLen
					bestRunStart = currentRunStart
				}
				currentRunLen = 0
			}
		}
	}
	if currentRunLen > bestRunLen {
		bestRunLen = currentRunLen
		bestRunStart = currentRunStart
	}

	// Minimum 15 seconds of matching audio (~121 samples)
	minSamples := int(math.Round(15.0 * SamplesPerSecond))
	if bestRunLen < minSamples {
		return nil
	}

	// Convert sample positions to seconds
	secPerSample := 1.0 / SamplesPerSecond
	matchStartA := float64(startA+bestRunStart) * secPerSample
	matchEndA := float64(startA+bestRunStart+bestRunLen) * secPerSample
	matchStartB := float64(startB+bestRunStart) * secPerSample
	matchEndB := float64(startB+bestRunStart+bestRunLen) * secPerSample

	// Calculate match quality score
	score := calculateMatchScore(a, b, startA+bestRunStart, startB+bestRunStart, bestRunLen, maxBitDiff)

	return &MatchSegment{
		StartA: matchStartA,
		EndA:   matchEndA,
		StartB: matchStartB,
		EndB:   matchEndB,
		Score:  score,
	}
}

// calculateMatchScore computes a 0.0-1.0 quality score for a matched region.
func calculateMatchScore(a, b []uint32, startA, startB, length, maxBitDiff int) float64 {
	if length == 0 {
		return 0
	}
	totalBits := 0
	for i := 0; i < length; i++ {
		totalBits += bits.OnesCount32(a[startA+i] ^ b[startB+i])
	}
	avgBitDiff := float64(totalBits) / float64(length)
	// Score: 1.0 when avgBitDiff=0, 0.0 when avgBitDiff=maxBitDiff
	score := 1.0 - (avgBitDiff / float64(maxBitDiff))
	return math.Max(0, math.Min(1, score))
}

// FingerprintToBytes converts uint32 fingerprint to little-endian byte slice for DB storage.
func FingerprintToBytes(fp []uint32) []byte {
	buf := make([]byte, len(fp)*4)
	for i, v := range fp {
		binary.LittleEndian.PutUint32(buf[i*4:], v)
	}
	return buf
}

// BytesToFingerprint converts little-endian byte slice back to uint32 fingerprint.
func BytesToFingerprint(data []byte) []uint32 {
	if len(data)%4 != 0 {
		return nil
	}
	fp := make([]uint32, len(data)/4)
	for i := range fp {
		fp[i] = binary.LittleEndian.Uint32(data[i*4:])
	}
	return fp
}

// FindSeasonIntro compares fingerprints across episodes to find shared intro.
// fingerprints maps media_file_id → uint32 array.
// Returns the consensus intro region or nil if not enough episodes match.
func FindSeasonIntro(fingerprints map[int64][]uint32, minEpisodes int, maxBitDiff int) *SeasonIntroResult {
	if len(fingerprints) < 2 {
		return nil
	}

	// Collect file IDs for ordered iteration
	fileIDs := make([]int64, 0, len(fingerprints))
	for id := range fingerprints {
		fileIDs = append(fileIDs, id)
	}

	// Use first file as reference, compare against others
	refID := fileIDs[0]
	refFP := fingerprints[refID]

	var matches []fileMatch
	for _, otherID := range fileIDs[1:] {
		seg := CompareFingerprints(refFP, fingerprints[otherID], maxBitDiff)
		if seg != nil {
			matches = append(matches, fileMatch{
				fileID:  otherID,
				segment: seg,
			})
		}
	}

	// Need enough matching episodes
	if len(matches)+1 < minEpisodes { // +1 for reference
		return nil
	}

	// Use the reference file's start/end as the consensus
	// (all matches are relative to the reference)
	var totalStart, totalEnd float64
	var totalScore float64
	for _, m := range matches {
		totalStart += m.segment.StartA
		totalEnd += m.segment.EndA
		totalScore += m.segment.Score
	}

	n := float64(len(matches))
	avgStart := totalStart / n
	avgEnd := totalEnd / n
	avgScore := totalScore / n

	// Build per-file results
	perFile := make(map[int64]IntroTimestamps)
	perFile[refID] = IntroTimestamps{Start: avgStart, End: avgEnd}
	for _, m := range matches {
		perFile[m.fileID] = IntroTimestamps{Start: m.segment.StartB, End: m.segment.EndB}
	}

	return &SeasonIntroResult{
		Start:      avgStart,
		End:        avgEnd,
		Confidence: 0.5 + avgScore*0.4, // Map score 0-1 to confidence 0.5-0.9
		MatchCount: len(matches) + 1,
		PerFile:    perFile,
	}
}

// SeasonIntroResult holds the consensus intro detection for a season.
type SeasonIntroResult struct {
	Start      float64                   // consensus start (seconds)
	End        float64                   // consensus end (seconds)
	Confidence float64                   // 0.5-0.9
	MatchCount int                       // number of episodes that matched
	PerFile    map[int64]IntroTimestamps // per-file timestamps (file_id → timestamps)
}

// IntroTimestamps holds per-file intro start/end.
type IntroTimestamps struct {
	Start float64
	End   float64
}

type fileMatch struct {
	fileID  int64
	segment *MatchSegment
}
