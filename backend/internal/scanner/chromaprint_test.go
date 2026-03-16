package scanner

import (
	"testing"
)

func TestParseFpcalcOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid output",
			output:  "DURATION=120\nFINGERPRINT=123456,789012,345678\n",
			wantLen: 3,
		},
		{
			name:    "negative values (signed int32)",
			output:  "DURATION=60\nFINGERPRINT=-1234567,789012,-345678\n",
			wantLen: 3,
		},
		{
			name:    "empty fingerprint",
			output:  "DURATION=0\nFINGERPRINT=\n",
			wantLen: 0,
		},
		{
			name:    "no fingerprint line",
			output:  "DURATION=120\n",
			wantErr: true,
		},
		{
			name:    "single value",
			output:  "FINGERPRINT=42\n",
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp, err := parseFpcalcOutput(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fp) != tt.wantLen {
				t.Errorf("got %d samples, want %d", len(fp), tt.wantLen)
			}
		})
	}
}

func TestFingerprintBytesRoundtrip(t *testing.T) {
	original := []uint32{0, 1, 42, 0xFFFFFFFF, 0xDEADBEEF}
	bytes := FingerprintToBytes(original)
	restored := BytesToFingerprint(bytes)

	if len(restored) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(restored), len(original))
	}
	for i := range original {
		if restored[i] != original[i] {
			t.Errorf("index %d: got %d, want %d", i, restored[i], original[i])
		}
	}
}

func TestCompareFingerprints_IdenticalAudio(t *testing.T) {
	// Two identical fingerprints should match perfectly
	n := 500
	fp := make([]uint32, n)
	for i := range fp {
		fp[i] = uint32(i * 12345)
	}

	seg := CompareFingerprints(fp, fp, 10)
	if seg == nil {
		t.Fatal("expected match for identical fingerprints")
	}
	if seg.Score < 0.9 {
		t.Errorf("score = %.2f, expected > 0.9 for identical audio", seg.Score)
	}
}

func TestCompareFingerprints_ShiftedMatch(t *testing.T) {
	// Same audio but shifted by 50 samples (~6 seconds)
	n := 500
	shared := make([]uint32, 300)
	for i := range shared {
		shared[i] = uint32(i*7777 + 42)
	}

	a := make([]uint32, n)
	b := make([]uint32, n)

	// Place shared audio at different offsets
	offsetA := 10
	offsetB := 60
	copy(a[offsetA:], shared)
	copy(b[offsetB:], shared)

	// Fill non-shared regions with different noise
	for i := 0; i < offsetA; i++ {
		a[i] = uint32(i * 99999)
	}
	for i := 0; i < offsetB; i++ {
		b[i] = uint32(i * 88888)
	}

	seg := CompareFingerprints(a, b, 10)
	if seg == nil {
		t.Fatal("expected match for shifted identical audio")
	}
	if seg.Score < 0.8 {
		t.Errorf("score = %.2f, expected > 0.8", seg.Score)
	}
}

func TestCompareFingerprints_NoMatch(t *testing.T) {
	n := 200
	a := make([]uint32, n)
	b := make([]uint32, n)
	for i := range a {
		a[i] = uint32(i * 11111)
		b[i] = uint32(i*99999 + 77777)
	}

	seg := CompareFingerprints(a, b, 10)
	if seg != nil {
		t.Errorf("expected no match for completely different audio, got %+v", seg)
	}
}

func TestCompareFingerprints_TooShort(t *testing.T) {
	a := []uint32{1, 2, 3}
	b := []uint32{1, 2, 3}
	seg := CompareFingerprints(a, b, 10)
	if seg != nil {
		t.Error("expected nil for too-short fingerprints")
	}
}

func TestFindSeasonIntro(t *testing.T) {
	// Simulate 4 episodes with shared intro (first 200 samples)
	introPattern := make([]uint32, 200)
	for i := range introPattern {
		introPattern[i] = uint32(i*5555 + 1234)
	}

	// Use a simple hash to generate non-overlapping content per episode
	hash := func(seed, i int) uint32 {
		v := uint32(seed*2654435761 + i*1103515245)
		return v ^ (v >> 16) // Mix bits to avoid systematic patterns
	}

	fingerprints := make(map[int64][]uint32)
	for fileID := int64(1); fileID <= 4; fileID++ {
		fp := make([]uint32, 500)
		copy(fp, introPattern)
		// Fill rest with truly different per-episode content
		for i := 200; i < 500; i++ {
			fp[i] = hash(int(fileID)*77777, i)
		}
		fingerprints[fileID] = fp
	}

	result := FindSeasonIntro(fingerprints, 2, 10)
	if result == nil {
		t.Fatal("expected season intro detection")
	}
	if result.MatchCount < 2 {
		t.Errorf("match count = %d, expected >= 2", result.MatchCount)
	}
	if result.Confidence < 0.5 {
		t.Errorf("confidence = %.2f, expected >= 0.5", result.Confidence)
	}
	// Intro should end around sample 200 → ~24.7 seconds
	expectedEnd := 200.0 / SamplesPerSecond
	if result.End < expectedEnd*0.7 || result.End > expectedEnd*1.3 {
		t.Errorf("end = %.1f, expected ~%.1f (±30%%)", result.End, expectedEnd)
	}
}

func TestFindSeasonIntro_NotEnoughEpisodes(t *testing.T) {
	fingerprints := map[int64][]uint32{
		1: make([]uint32, 100),
	}
	result := FindSeasonIntro(fingerprints, 2, 10)
	if result != nil {
		t.Error("expected nil with only 1 episode")
	}
}
