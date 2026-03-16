package scanner

import (
	"context"

	"github.com/thawng/velox/internal/model"
)

// MarkerDetector is the interface for different marker detection strategies
type MarkerDetector interface {
	// Detect attempts to find intro/credits markers for a media file
	// Returns a list of detected markers (may be empty if none found)
	// Returns error if detection fails
	Detect(ctx context.Context, fileID int64, filePath string) ([]DetectedMarker, error)

	// Name returns the detector identifier (e.g., "chapter", "fingerprint", "manual")
	Name() string

	// Confidence returns the default confidence level for this detector's results
	// chapter = 1.0, fingerprint < 1.0, manual = 1.0
	Confidence() float64
}

// DetectedMarker represents a marker found by a detector
type DetectedMarker struct {
	Type       string // "intro" or "credits"
	StartSec   float64
	EndSec     float64
	Label      string // Optional label (e.g., original chapter title)
	Source     string // Detector name
	Confidence float64
}

// ToModel converts a DetectedMarker to a model.MediaMarker
func (dm DetectedMarker) ToModel(fileID int64) *model.MediaMarker {
	return &model.MediaMarker{
		MediaFileID: fileID,
		MarkerType:  dm.Type,
		StartSec:    dm.StartSec,
		EndSec:      dm.EndSec,
		Source:      dm.Source,
		Confidence:  dm.Confidence,
		Label:       dm.Label,
	}
}

// DetectorRegistry holds all available marker detectors
type DetectorRegistry struct {
	detectors map[string]MarkerDetector
}

// NewDetectorRegistry creates a new registry
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{
		detectors: make(map[string]MarkerDetector),
	}
}

// Register adds a detector to the registry
func (r *DetectorRegistry) Register(detector MarkerDetector) {
	r.detectors[detector.Name()] = detector
}

// Get retrieves a detector by name
func (r *DetectorRegistry) Get(name string) (MarkerDetector, bool) {
	d, ok := r.detectors[name]
	return d, ok
}

// All returns all registered detectors
func (r *DetectorRegistry) All() []MarkerDetector {
	result := make([]MarkerDetector, 0, len(r.detectors))
	for _, d := range r.detectors {
		result = append(result, d)
	}
	return result
}

// SourcePriority defines the priority order for marker sources
// Higher number = higher priority
var SourcePriority = map[string]int{
	"manual":      3,
	"chapter":     2,
	"fingerprint": 1,
}

// CompareSourcePriority returns true if sourceA has higher priority than sourceB
func CompareSourcePriority(sourceA, sourceB string) bool {
	priorityA := SourcePriority[sourceA]
	priorityB := SourcePriority[sourceB]
	return priorityA > priorityB
}
