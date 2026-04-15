package model

import (
	"encoding/json"
	"testing"
)

func TestImagePathNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"local media", "local://media/42/poster.jpg", "/api/images/local/media/42/poster.jpg"},
		{"local series", "local://series/7/backdrop.jpg", "/api/images/local/series/7/backdrop.jpg"},
		{"local episode still", "local://episode/99/still.jpg", "/api/images/local/episode/99/still.jpg"},
		// Untyped ImagePath leaves raw TMDb paths unchanged — callers must use
		// a typed wrapper (PosterPath/BackdropPath/ProfilePath/…) to route
		// through /api/images/tmdb/{type}/{path}.
		{"tmdb relative untyped", "/abc123.jpg", "/abc123.jpg"},
		{"already normalized", "/api/images/local/media/1/poster.jpg", "/api/images/local/media/1/poster.jpg"},
		{"external https", "https://image.tmdb.org/t/p/w500/abc.jpg", "https://image.tmdb.org/t/p/w500/abc.jpg"},
		{"external http", "http://cdn.example.com/x.png", "http://cdn.example.com/x.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImagePath(tt.in).Normalize(); got != tt.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestImagePathMarshalJSONIdempotent(t *testing.T) {
	p := ImagePath("local://media/1/poster.jpg")
	first, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Re-unmarshal then re-marshal: normalized value must round-trip unchanged.
	var back ImagePath
	if err := json.Unmarshal(first, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	second, err := json.Marshal(back)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("normalization not idempotent: first=%s second=%s", first, second)
	}
}

func TestTypedPathNormalize(t *testing.T) {
	if got := PosterPath("/abc.jpg").Normalize(); got != "/api/images/tmdb/poster/abc.jpg" {
		t.Fatalf("unexpected poster normalize: %s", got)
	}
	if got := BackdropPath("/abc.jpg").Normalize(); got != "/api/images/tmdb/backdrop/abc.jpg" {
		t.Fatalf("unexpected backdrop normalize: %s", got)
	}
	if got := StillPath("/abc.jpg").Normalize(); got != "/api/images/tmdb/still/abc.jpg" {
		t.Fatalf("unexpected still normalize: %s", got)
	}
	if got := LogoPath("/abc.jpg").Normalize(); got != "/api/images/tmdb/logo/abc.jpg" {
		t.Fatalf("unexpected logo normalize: %s", got)
	}
	if got := ProfilePath("/abc.jpg").Normalize(); got != "/api/images/tmdb/profile/abc.jpg" {
		t.Fatalf("unexpected profile normalize: %s", got)
	}
	// local:// and /api/... forms must pass through regardless of type wrapper.
	if got := ProfilePath("local://user/7/avatar.jpg").Normalize(); got != "/api/images/local/user/7/avatar.jpg" {
		t.Fatalf("unexpected profile local normalize: %s", got)
	}
}
