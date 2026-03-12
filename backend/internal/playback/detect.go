package playback

import (
	"regexp"
	"strings"
)

// ClientCapabilities represents detected client capabilities
type ClientCapabilities struct {
	Browser  string         `json:"browser"`
	Platform string         `json:"platform"`
	IsMobile bool           `json:"is_mobile"`
	Profile  *DeviceProfile `json:"profile"`
}

// User agent patterns for detection
var (
	// Browser patterns
	chromePattern  = regexp.MustCompile(`(?i)chrome|chromium|crios`)
	firefoxPattern = regexp.MustCompile(`(?i)firefox|fxios`)
	safariPattern  = regexp.MustCompile(`(?i)safari`)
	edgePattern    = regexp.MustCompile(`(?i)edg`)

	// Platform patterns
	windowsPattern = regexp.MustCompile(`(?i)windows`)
	macOSPattern   = regexp.MustCompile(`(?i)macintosh|mac os x`)
	linuxPattern   = regexp.MustCompile(`(?i)linux`)
	iosPattern     = regexp.MustCompile(`(?i)iphone|ipad|ipod`)
	androidPattern = regexp.MustCompile(`(?i)android`)

	// Mobile detection
	mobilePattern = regexp.MustCompile(`(?i)mobile|android|iphone|ipad|ipod`)
)

// DetectClientFromUA detects browser and platform from User-Agent string
func DetectClientFromUA(userAgent string) ClientCapabilities {
	caps := ClientCapabilities{
		Browser:  "unknown",
		Platform: "unknown",
		Profile:  &GenericBrowser,
	}

	ua := strings.ToLower(userAgent)

	// Detect platform
	if iosPattern.MatchString(ua) {
		caps.Platform = "ios"
		caps.IsMobile = true
	} else if androidPattern.MatchString(ua) {
		caps.Platform = "android"
		caps.IsMobile = true
	} else if windowsPattern.MatchString(ua) {
		caps.Platform = "windows"
	} else if macOSPattern.MatchString(ua) {
		caps.Platform = "macos"
	} else if linuxPattern.MatchString(ua) {
		caps.Platform = "linux"
	}

	// Detect browser (order matters - check Edge before Chrome)
	if edgePattern.MatchString(ua) {
		caps.Browser = "edge"
		caps.Profile = &EdgeDesktop
	} else if chromePattern.MatchString(ua) {
		caps.Browser = "chrome"
		if caps.Platform == "ios" {
			// Chrome on iOS is actually Safari WebKit
			caps.Browser = "safari"
			caps.Profile = &MobileSafari
		} else {
			caps.Profile = &ChromeDesktop
		}
	} else if firefoxPattern.MatchString(ua) {
		caps.Browser = "firefox"
		caps.Profile = &FirefoxDesktop
	} else if safariPattern.MatchString(ua) {
		caps.Browser = "safari"
		if caps.IsMobile || caps.Platform == "ios" {
			caps.Profile = &MobileSafari
		} else {
			caps.Profile = &SafariDesktop
		}
	}

	// Adjust for mobile
	if caps.IsMobile && caps.Profile != &MobileSafari {
		// Use generic mobile profile for unknown mobile browsers
		mobileProfile := *caps.Profile
		mobileProfile.MaxHeight = 1080
		mobileProfile.MaxBitrate = 15000
		caps.Profile = &mobileProfile
	}

	return caps
}

// DetectClient detects client from request headers
func DetectClient(userAgent string) *DeviceProfile {
	caps := DetectClientFromUA(userAgent)
	return caps.Profile
}

// GetClientInfo returns detailed client info for API responses
func GetClientInfo(userAgent string) ClientCapabilities {
	return DetectClientFromUA(userAgent)
}

// IsHLSPreferred checks if client prefers HLS over direct file
func (c *ClientCapabilities) IsHLSPreferred() bool {
	if c.Profile == nil {
		return false
	}
	return c.Profile.SupportsHLS && c.IsMobile
}

// NeedsTranscodeForResolution determines if transcoding is needed for given resolution
func (c *ClientCapabilities) NeedsTranscodeForResolution(width, height int) bool {
	if c.Profile == nil {
		return false
	}
	return !c.Profile.CanPlayResolution(width, height)
}
