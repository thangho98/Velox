package model

import (
	"testing"
)

func TestBuildImageResource(t *testing.T) {
	tests := []struct {
		name     string
		rawPath  string
		kind     string
		meta     *ImageMetadata
		wantURL  string
		wantType string
		wantAsp  string
	}{
		{
			name:     "poster tmdb path",
			rawPath:  "/fsdf234.jpg",
			kind:     "poster",
			meta:     nil,
			wantURL:  "/api/images/tmdb/poster/fsdf234.jpg?width=500",
			wantType: "poster",
			wantAsp:  "2:3",
		},
		{
			name:     "backdrop with meta",
			rawPath:  "/bd.jpg",
			kind:     "backdrop",
			meta:     &ImageMetadata{Width: 1920, Height: 1080, Blurhash: "hash123"},
			wantURL:  "/api/images/tmdb/backdrop/bd.jpg?width=780",
			wantType: "backdrop",
			wantAsp:  "16:9",
		},
		{
			name:     "local path",
			rawPath:  "local:///path/to/image.jpg",
			kind:     "poster",
			meta:     nil,
			wantURL:  "/api/images/local//path/to/image.jpg?width=500",
			wantType: "poster",
			wantAsp:  "2:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := BuildImageResource(tt.rawPath, tt.kind, tt.meta)
			if res == nil {
				if tt.rawPath != "" {
					t.Fatalf("expected resource, got nil")
				}
				return
			}
			if res.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", res.URL, tt.wantURL)
			}
			if res.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", res.Type, tt.wantType)
			}
			if res.Aspect != tt.wantAsp {
				t.Errorf("Aspect = %q, want %q", res.Aspect, tt.wantAsp)
			}
			if tt.meta != nil {
				if res.Width != tt.meta.Width || res.Height != tt.meta.Height {
					t.Errorf("dimensions = %dx%d, want %dx%d", res.Width, res.Height, tt.meta.Width, tt.meta.Height)
				}
				if res.Blurhash == nil || *res.Blurhash != tt.meta.Blurhash {
					t.Errorf("blurhash mismatch")
				}
			} else {
				if res.Blurhash != nil {
					t.Errorf("expected nil blurhash, got %v", *res.Blurhash)
				}
			}
		})
	}
}
