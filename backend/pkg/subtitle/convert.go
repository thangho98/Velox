package subtitle

import "strings"

// SRTToVTT converts SRT subtitle content to WebVTT format.
// Handles BOM, CRLF line endings, and the SRT timestamp separator (comma → dot).
func SRTToVTT(data []byte) []byte {
	content := strings.TrimPrefix(string(data), "\xef\xbb\xbf") // strip BOM
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, " --> ") {
			parts := strings.SplitN(line, " --> ", 2)
			if len(parts) == 2 {
				lines[i] = strings.ReplaceAll(parts[0], ",", ".") +
					" --> " +
					strings.ReplaceAll(parts[1], ",", ".")
			}
		}
	}
	return []byte("WEBVTT\n\n" + strings.Join(lines, "\n"))
}
