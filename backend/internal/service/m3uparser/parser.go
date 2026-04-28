package m3uparser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/thawng/velox/internal/model"
)

var (
	// Regex matches key="value" pairs in the EXTINF line
	attrRegex = regexp.MustCompile(`([a-zA-Z0-9_-]+)="([^"]*)"`)

	// #EXTVLCOPT and #KODIPROP carry per-channel HTTP hints that origin CDNs
	// require (e.g. MyTV expects User-Agent "MyTV"). We extract UA/Referer/Origin
	// and persist them so the proxy can forward them upstream.
	extvlcoptUARegex      = regexp.MustCompile(`(?i)http-user-agent\s*=\s*"?([^"\r\n]+?)"?\s*$`)
	extvlcoptRefererRegex = regexp.MustCompile(`(?i)http-referrer\s*=\s*"?([^"\r\n]+?)"?\s*$`)
	extvlcoptOriginRegex  = regexp.MustCompile(`(?i)http-origin\s*=\s*"?([^"\r\n]+?)"?\s*$`)
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

// ParseURL downloads and parses an M3U playlist from a URL with context cancellation
func (p *Parser) ParseURL(ctx context.Context, playlistURL string) ([]*model.LiveChannel, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request failed: %w", err)
	}

	client := &http.Client{} // Or pass it in Parser struct if needed, for now Default-like client
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch playlist URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("unexpected HTTP status fetching playlist: %s", resp.Status)
	}

	return p.Parse(resp.Body, playlistURL)
}

// ParseFile reads and parses an M3U playlist from a local file
func (p *Parser) ParseFile(filepath string) ([]*model.LiveChannel, string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	// Use filepath as base to normalize relative paths if needed, though they'd be local file paths
	return p.Parse(file, filepath)
}

// extractCountryFromURL infers country code from iptv-org style URLs
func extractCountryFromURL(u string) string {
	if strings.Contains(u, "/countries/") {
		parts := strings.Split(u, "/")
		last := parts[len(parts)-1]
		if strings.HasSuffix(last, ".m3u") {
			countryCode := strings.TrimSuffix(last, ".m3u")
			if len(countryCode) == 2 {
				return strings.ToUpper(countryCode)
			}
		}
	}
	return ""
}

// Parse processes an io.Reader containing M3U data and returns a slice of LiveChannels and an optional EpgURL
func (p *Parser) Parse(r io.Reader, baseURL string) ([]*model.LiveChannel, string, error) {
	scanner := bufio.NewScanner(r)

	// Increase scanner buffer size to handle extremely long lines in some weird M3U files
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var channels []*model.LiveChannel
	var currentChannel *model.LiveChannel
	var epgURL string

	var base *url.URL
	if baseURL != "" {
		base, _ = url.Parse(baseURL)
	}
	defaultCountry := extractCountryFromURL(baseURL)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// EXTINF line contains metadata
		if strings.HasPrefix(line, "#EXTINF:") {
			currentChannel = parseExtInf(line)
			continue
		}

		// Handle the EXTM3U header tag specifically for url-tvg or x-tvg-url
		if strings.HasPrefix(line, "#EXTM3U") {
			matches := attrRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) == 3 {
					key := strings.ToLower(match[1])
					if key == "url-tvg" || key == "x-tvg-url" {
						epgURL = match[2]
					}
				}
			}
			continue
		}

		// Per-channel HTTP overrides: #EXTVLCOPT:http-user-agent="..." etc.
		// Some repos use #KODIPROP:inputstream.adaptive.manifest_headers too,
		// but the common/portable form we honor here is EXTVLCOPT.
		if currentChannel != nil && (strings.HasPrefix(line, "#EXTVLCOPT") || strings.HasPrefix(line, "#KODIPROP")) {
			if m := extvlcoptUARegex.FindStringSubmatch(line); len(m) == 2 && currentChannel.UserAgent == "" {
				currentChannel.UserAgent = strings.TrimSpace(m[1])
			}
			if m := extvlcoptRefererRegex.FindStringSubmatch(line); len(m) == 2 && currentChannel.Referer == "" {
				currentChannel.Referer = strings.TrimSpace(m[1])
			}
			if m := extvlcoptOriginRegex.FindStringSubmatch(line); len(m) == 2 && currentChannel.Origin == "" {
				currentChannel.Origin = strings.TrimSpace(m[1])
			}
			continue
		}

		// If it's a comment but not EXTINF, ignore it
		if strings.HasPrefix(line, "#") {
			continue
		}

		// It must be a Stream URL line
		if currentChannel != nil {
			streamURLStr := line

			// Normalize relative URLs against the base playlist URL
			if base != nil {
				parsedLine, err := url.Parse(line)
				if err == nil && parsedLine.Scheme == "" && parsedLine.Host == "" {
					streamURLStr = base.ResolveReference(parsedLine).String()
				}
			}

			currentChannel.StreamURL = streamURLStr
			currentChannel.IsActive = true

			// Infer country if missing
			if currentChannel.Country == "" && defaultCountry != "" {
				currentChannel.Country = defaultCountry
			}

			channels = append(channels, currentChannel)
			currentChannel = nil // Reset for next channel
		}
	}

	if err := scanner.Err(); err != nil {
		return channels, epgURL, err
	}

	return channels, epgURL, nil
}

// parseExtInf extracts attributes from the #EXTINF line
// Example: #EXTINF:-1 tvg-id="vtv1" tvg-logo="url" group-title="News",VTV1 HD
func parseExtInf(line string) *model.LiveChannel {
	channel := &model.LiveChannel{}

	// Locate the separating comma. It's the first comma that isn't inside quotes.
	inQuotes := false
	commaIdx := -1
	for i, ch := range line {
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			commaIdx = i
			break
		}
	}

	var attrPart string
	if commaIdx != -1 && commaIdx < len(line) {
		channel.Name = strings.TrimSpace(line[commaIdx+1:])
		attrPart = line[:commaIdx]
	} else {
		channel.Name = "Unknown Channel"
		attrPart = line
	}

	// Find all attributes using regex
	matches := attrRegex.FindAllStringSubmatch(attrPart, -1)
	for _, match := range matches {
		if len(match) == 3 {
			key := strings.ToLower(match[1])
			val := match[2]

			switch key {
			case "tvg-id":
				channel.ChannelID = val
			case "tvg-logo":
				channel.Logo = val
			case "group-title":
				channel.GroupTitle = val
			case "tvg-country":
				channel.Country = strings.ToUpper(val)
			}
		}
	}

	return channel
}
