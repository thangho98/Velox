package handler

import (
	"testing"
)

func TestResolveTMDbSize(t *testing.T) {
	tests := []struct {
		imageType string
		width     string
		want      string
		ok        bool
	}{
		{"poster", "500", "w500", true},
		{"poster", "400", "w500", true},
		{"poster", "0", "w500", true},
		{"poster", "invalid", "w500", true},
		{"poster", "original", "original", true},
		{"poster", "9999", "original", true},

		{"backdrop", "780", "w780", true},
		{"backdrop", "500", "w780", true},
		{"backdrop", "0", "w780", true},

		{"still", "185", "w185", true},
		{"logo", "154", "w154", true},

		{"unknown", "500", "", false},
	}

	for _, tt := range tests {
		got, ok := resolveTMDbSize(tt.imageType, tt.width)
		if ok != tt.ok || got != tt.want {
			t.Errorf("resolveTMDbSize(%q, %q) = %q, %v; want %q, %v", tt.imageType, tt.width, got, ok, tt.want, tt.ok)
		}
	}
}
