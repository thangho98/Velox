package fshare

import (
	"errors"
	"testing"

	"github.com/thawng/velox/internal/cloudstorage"
)

func TestParseFolderURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		err  error
	}{
		{"www https", "https://www.fshare.vn/folder/XZWCPAZV3J71", "XZWCPAZV3J71", nil},
		{"bare host", "https://fshare.vn/folder/ABC123", "ABC123", nil},
		{"trailing path", "https://www.fshare.vn/folder/XYZ/foo/bar", "XYZ", nil},
		{"query params", "https://www.fshare.vn/folder/XYZ?utm_source=x", "XYZ", nil},
		{"underscore+dash", "https://www.fshare.vn/folder/ab-CD_12", "ab-CD_12", nil},
		{"whitespace trimmed", "   https://www.fshare.vn/folder/TRIM   ", "TRIM", nil},

		{"empty", "", "", cloudstorage.ErrInvalidURL},
		{"wrong host", "https://google.com/folder/XYZ", "", cloudstorage.ErrInvalidURL},
		{"file path", "https://www.fshare.vn/file/XYZ", "", cloudstorage.ErrInvalidURL},
		{"no code", "https://www.fshare.vn/folder/", "", cloudstorage.ErrInvalidURL},
		{"garbage", "not a url", "", cloudstorage.ErrInvalidURL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFolderURL(tc.in)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Errorf("err = %v; want %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestParseFileURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		err  error
	}{
		{"www https", "https://www.fshare.vn/file/ABC12345", "ABC12345", nil},
		{"bare host", "https://fshare.vn/file/XYZ", "XYZ", nil},
		{"query params", "https://www.fshare.vn/file/XYZ?t=1", "XYZ", nil},

		{"folder path", "https://www.fshare.vn/folder/XYZ", "", cloudstorage.ErrInvalidURL},
		{"wrong host", "https://other.com/file/XYZ", "", cloudstorage.ErrInvalidURL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFileURL(tc.in)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Errorf("err = %v; want %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}
