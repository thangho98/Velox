package scanner

import (
	"testing"
)

func TestParseBlackFrameOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
	}{
		{
			name: "typical output",
			output: "[blackframe @ 0x1234] frame:100 pblack:98 pts:250000 t:10.000 type:I last_keyframe:100\n" +
				"[blackframe @ 0x1234] frame:200 pblack:99 pts:500000 t:20.500 type:P last_keyframe:100",
			wantLen: 2,
		},
		{
			name:    "no black frames",
			output:  "frame=1000 fps=60.0 q=-1.0 size=N/A time=00:00:10.00 bitrate=N/A speed=1.00x",
			wantLen: 0,
		},
		{
			name:    "empty output",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "single frame",
			output:  "[blackframe @ 0xabc] frame:50 pblack:95 pts:125000 t:5.250 type:I last_keyframe:50",
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frames := parseBlackFrameOutput(tt.output)
			if len(frames) != tt.wantLen {
				t.Errorf("got %d frames, want %d", len(frames), tt.wantLen)
			}
		})
	}
}

func TestParseSilenceDetectOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
	}{
		{
			name: "two silence ranges",
			output: `[silencedetect @ 0x1234] silence_start: 120.500
[silencedetect @ 0x1234] silence_end: 122.000 | silence_duration: 1.500
[silencedetect @ 0x1234] silence_start: 180.000
[silencedetect @ 0x1234] silence_end: 181.500 | silence_duration: 1.500`,
			wantLen: 2,
		},
		{
			name:    "no silence",
			output:  "frame=1000 fps=60.0 time=00:00:10.00",
			wantLen: 0,
		},
		{
			name: "trailing silence_start without end",
			output: `[silencedetect @ 0x1234] silence_start: 120.500
[silencedetect @ 0x1234] silence_end: 122.000 | silence_duration: 1.500
[silencedetect @ 0x1234] silence_start: 590.000`,
			wantLen: 2, // 1 complete pair + 1 trailing
		},
		{
			name:    "empty",
			output:  "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges := parseSilenceDetectOutput(tt.output)
			if len(ranges) != tt.wantLen {
				t.Errorf("got %d ranges, want %d", len(ranges), tt.wantLen)
			}
		})
	}
}

func TestFindCreditsBoundary(t *testing.T) {
	tests := []struct {
		name           string
		blackFrames    []float64
		silences       []SilenceRange
		regionStart    float64
		totalDuration  float64
		wantStart      float64
		wantConfidence float64
	}{
		{
			name:           "black frame + silence co-occur",
			blackFrames:    []float64{500, 501, 502, 503, 504}, // cluster at t=500 (abs: 1700)
			silences:       []SilenceRange{{Start: 498, End: 500}},
			regionStart:    1200,
			totalDuration:  2000,
			wantStart:      1700, // regionStart + 500
			wantConfidence: 0.75,
		},
		{
			name:           "black frame only (no silence)",
			blackFrames:    []float64{400, 401, 402, 403},
			silences:       nil,
			regionStart:    1200,
			totalDuration:  2000,
			wantStart:      1600, // regionStart + 400
			wantConfidence: 0.55,
		},
		{
			name:           "no black frames, no silence",
			blackFrames:    nil,
			silences:       nil,
			regionStart:    1200,
			totalDuration:  2000,
			wantStart:      0,
			wantConfidence: 0,
		},
		{
			name:           "cluster too early (< 70% of total)",
			blackFrames:    []float64{10, 11, 12, 13},
			silences:       []SilenceRange{{Start: 9, End: 11}},
			regionStart:    0,
			totalDuration:  2000,
			wantStart:      0, // Too early to be credits
			wantConfidence: 0,
		},
		{
			name:           "not enough black frames for cluster",
			blackFrames:    []float64{500, 501}, // Only 2, need 3
			silences:       []SilenceRange{{Start: 499, End: 501}},
			regionStart:    1200,
			totalDuration:  2000,
			wantStart:      0,
			wantConfidence: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, confidence := findCreditsBoundary(tt.blackFrames, tt.silences, tt.regionStart, tt.totalDuration)
			if tt.wantStart == 0 {
				if start != 0 {
					t.Errorf("expected no boundary, got start=%.1f", start)
				}
				return
			}
			if start < tt.wantStart-5 || start > tt.wantStart+5 {
				t.Errorf("start = %.1f, want ~%.1f", start, tt.wantStart)
			}
			if confidence != tt.wantConfidence {
				t.Errorf("confidence = %.2f, want %.2f", confidence, tt.wantConfidence)
			}
		})
	}
}

func TestFindBlackFrameClusters(t *testing.T) {
	tests := []struct {
		name     string
		frames   []float64
		window   float64
		minCount int
		wantLen  int
	}{
		{
			name:     "one cluster of 5",
			frames:   []float64{100, 101, 102, 103, 104},
			window:   10,
			minCount: 3,
			wantLen:  1,
		},
		{
			name:     "two clusters",
			frames:   []float64{100, 101, 102, 200, 201, 202},
			window:   10,
			minCount: 3,
			wantLen:  2,
		},
		{
			name:     "spread out, no cluster",
			frames:   []float64{100, 200, 300},
			window:   10,
			minCount: 3,
			wantLen:  0,
		},
		{
			name:     "empty",
			frames:   nil,
			window:   10,
			minCount: 3,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusters := findBlackFrameClusters(tt.frames, tt.window, tt.minCount)
			if len(clusters) != tt.wantLen {
				t.Errorf("got %d clusters, want %d", len(clusters), tt.wantLen)
			}
		})
	}
}
