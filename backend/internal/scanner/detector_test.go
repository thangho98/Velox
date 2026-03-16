package scanner

import (
	"context"
	"testing"
)

func TestCompareSourcePriority(t *testing.T) {
	tests := []struct {
		name     string
		sourceA  string
		sourceB  string
		expected bool
	}{
		{
			name:     "manual > chapter",
			sourceA:  "manual",
			sourceB:  "chapter",
			expected: true,
		},
		{
			name:     "chapter > fingerprint",
			sourceA:  "chapter",
			sourceB:  "fingerprint",
			expected: true,
		},
		{
			name:     "manual > fingerprint",
			sourceA:  "manual",
			sourceB:  "fingerprint",
			expected: true,
		},
		{
			name:     "chapter < manual",
			sourceA:  "chapter",
			sourceB:  "manual",
			expected: false,
		},
		{
			name:     "fingerprint < chapter",
			sourceA:  "fingerprint",
			sourceB:  "chapter",
			expected: false,
		},
		{
			name:     "same source - manual",
			sourceA:  "manual",
			sourceB:  "manual",
			expected: false,
		},
		{
			name:     "same source - fingerprint",
			sourceA:  "fingerprint",
			sourceB:  "fingerprint",
			expected: false,
		},
		{
			name:     "unknown source defaults to 0",
			sourceA:  "unknown",
			sourceB:  "fingerprint",
			expected: false,
		},
		{
			name:     "fingerprint > unknown",
			sourceA:  "fingerprint",
			sourceB:  "unknown",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareSourcePriority(tt.sourceA, tt.sourceB)
			if got != tt.expected {
				t.Errorf("CompareSourcePriority(%q, %q) = %v, want %v",
					tt.sourceA, tt.sourceB, got, tt.expected)
			}
		})
	}
}

func TestSourcePriorityValues(t *testing.T) {
	// Verify the expected priority values
	expected := map[string]int{
		"manual":      3,
		"chapter":     2,
		"fingerprint": 1,
	}

	for source, want := range expected {
		got, ok := SourcePriority[source]
		if !ok {
			t.Errorf("SourcePriority missing key: %q", source)
			continue
		}
		if got != want {
			t.Errorf("SourcePriority[%q] = %d, want %d", source, got, want)
		}
	}
}

func TestDetectedMarkerToModel(t *testing.T) {
	dm := DetectedMarker{
		Type:       "intro",
		StartSec:   10.5,
		EndSec:     85.0,
		Label:      "Opening",
		Source:     "chapter",
		Confidence: 1.0,
	}

	fileID := int64(42)
	model := dm.ToModel(fileID)

	if model.MediaFileID != fileID {
		t.Errorf("MediaFileID = %d, want %d", model.MediaFileID, fileID)
	}
	if model.MarkerType != "intro" {
		t.Errorf("MarkerType = %q, want %q", model.MarkerType, "intro")
	}
	if model.StartSec != 10.5 {
		t.Errorf("StartSec = %f, want %f", model.StartSec, 10.5)
	}
	if model.EndSec != 85.0 {
		t.Errorf("EndSec = %f, want %f", model.EndSec, 85.0)
	}
	if model.Source != "chapter" {
		t.Errorf("Source = %q, want %q", model.Source, "chapter")
	}
	if model.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want %f", model.Confidence, 1.0)
	}
	if model.Label != "Opening" {
		t.Errorf("Label = %q, want %q", model.Label, "Opening")
	}
}

func TestDetectorRegistry(t *testing.T) {
	registry := NewDetectorRegistry()

	// Create a mock detector
	mockDet := &mockDetector{
		name:       "mock",
		confidence: 0.5,
	}

	// Test Register and Get
	registry.Register(mockDet)

	got, ok := registry.Get("mock")
	if !ok {
		t.Fatal("expected to find 'mock' detector")
	}
	if got.Name() != "mock" {
		t.Errorf("Name() = %q, want %q", got.Name(), "mock")
	}

	// Test Get non-existent
	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("expected not to find 'nonexistent' detector")
	}

	// Test All
	all := registry.All()
	if len(all) != 1 {
		t.Errorf("len(All()) = %d, want 1", len(all))
	}

	// Register another detector
	mockDet2 := &mockDetector{
		name:       "mock2",
		confidence: 0.7,
	}
	registry.Register(mockDet2)

	all = registry.All()
	if len(all) != 2 {
		t.Errorf("len(All()) = %d, want 2", len(all))
	}
}

// mockDetector is a test helper that implements MarkerDetector
type mockDetector struct {
	name       string
	confidence float64
	detectFunc func(ctx context.Context, fileID int64, filePath string) ([]DetectedMarker, error)
}

func (m *mockDetector) Detect(ctx context.Context, fileID int64, filePath string) ([]DetectedMarker, error) {
	if m.detectFunc != nil {
		return m.detectFunc(ctx, fileID, filePath)
	}
	return nil, nil
}

func (m *mockDetector) Name() string {
	return m.name
}

func (m *mockDetector) Confidence() float64 {
	return m.confidence
}
