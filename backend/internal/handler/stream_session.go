package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const streamSessionQueryKey = "ssid"

func newStreamSessionID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func streamSessionIDFromValues(values url.Values) string {
	return sanitizeStreamSessionID(values.Get(streamSessionQueryKey))
}

func sanitizeStreamSessionID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if len(raw) > 64 {
		raw = raw[:64]
	}

	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		}
	}

	return b.String()
}

func extractStreamSessionIDFromSegmentName(name string) string {
	if !strings.HasPrefix(name, "ss") {
		return ""
	}
	rest := name[2:]
	idx := strings.IndexByte(rest, '_')
	if idx <= 0 {
		return ""
	}
	return sanitizeStreamSessionID(rest[:idx])
}
