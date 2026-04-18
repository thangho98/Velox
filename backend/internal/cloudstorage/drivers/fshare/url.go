package fshare

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/thawng/velox/internal/cloudstorage"
)

// Supported URL shapes:
//
//	https://www.fshare.vn/folder/<linkcode>
//	https://www.fshare.vn/folder/<linkcode>/trailing
//	https://fshare.vn/folder/<linkcode>
//	https://www.fshare.vn/file/<linkcode>
//	https://fshare.vn/file/<linkcode>?t=query&ignored
var (
	folderPathRe = regexp.MustCompile(`^/folder/([A-Za-z0-9_-]+)`)
	filePathRe   = regexp.MustCompile(`^/file/([A-Za-z0-9_-]+)`)
)

// parseFolderURL extracts a fshare folder linkcode from a user-pasted URL.
// Returns ErrInvalidURL with context when the URL isn't recognizable.
func parseFolderURL(raw string) (string, error) {
	return parseWithPattern(raw, folderPathRe, "folder")
}

// parseFileURL extracts a fshare file linkcode from a user-pasted URL.
func parseFileURL(raw string) (string, error) {
	return parseWithPattern(raw, filePathRe, "file")
}

func parseWithPattern(raw string, re *regexp.Regexp, kind string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%w: empty url", cloudstorage.ErrInvalidURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: parse: %v", cloudstorage.ErrInvalidURL, err)
	}
	host := strings.ToLower(u.Hostname())
	if !(host == "fshare.vn" || host == "www.fshare.vn") {
		return "", fmt.Errorf("%w: host %q is not fshare.vn", cloudstorage.ErrInvalidURL, host)
	}
	m := re.FindStringSubmatch(u.Path)
	if m == nil {
		return "", fmt.Errorf("%w: path %q does not match /%s/<linkcode>", cloudstorage.ErrInvalidURL, u.Path, kind)
	}
	return m[1], nil
}

// folderURL builds the canonical folder URL for a linkcode.
func folderURL(linkcode string) string { return "https://www.fshare.vn/folder/" + linkcode }

// fileURL builds the canonical file URL for a linkcode.
func fileURL(linkcode string) string { return "https://www.fshare.vn/file/" + linkcode }
