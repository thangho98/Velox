package playback

// AndroidNative profile for the Velox Android app using ExoPlayer/Media3.
// Codec/container support is further tightened by client-reported capabilities.
var AndroidNative = DeviceProfile{
	Name:                     "Android Native",
	SupportedVideoCodecs:     []string{CodecH264, CodecH265, CodecVP9, CodecAV1},
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecFLAC, CodecAC3, CodecEAC3},
	SupportedContainers:      []string{ContainerMP4, ContainerMKV, ContainerWebM, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT, SubtitleSRT, SubtitleASS},
	MaxWidth:                 0,
	MaxHeight:                0,
	MaxBitrate:               0,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             true,
	SupportsHDR:              true, // ExoPlayer/Media3 handles HDR rendering natively
	SupportsDolbyVision:      true, // Android handles Dolby Vision natively on capable devices
}

// VeloxDesktop profile for the Velox desktop client (Tauri 2 + libmpv).
// libmpv links jellyfin-ffmpeg with the SIMD `tonemapx` filter, which carries
// the IPT-PQ-C2 → BT.709 reshape natively. Client-side tonemap fires whenever
// the player observes `colormatrix=dolbyvision` and bypasses macOS's missing
// OpenGL 4.1 compute path, so DV P5 can direct-play here without server-side
// tonemapx burning NAS CPU. DV P7/8 keep direct-playing via the remove_dovi
// BSF (the enhancement layer is stripped before reaching the demuxer).
var VeloxDesktop = DeviceProfile{
	Name:                     "Velox Desktop (libmpv)",
	SupportedVideoCodecs:     []string{CodecH264, CodecH265, CodecVP9, CodecAV1},
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecFLAC, CodecAC3, CodecEAC3, CodecDTS, CodecTrueHD},
	SupportedContainers:      []string{ContainerMP4, ContainerMKV, ContainerWebM, ContainerHLS, ContainerMOV},
	SupportedSubtitleFormats: []string{SubtitleVTT, SubtitleSRT, SubtitleASS},
	MaxWidth:                 0,
	MaxHeight:                0,
	MaxBitrate:               0,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             true,
	SupportsHDR:              true,
	SupportsDolbyVision:      true,
}

// ChromeDesktop profile for Chrome/Chromium on desktop
var ChromeDesktop = DeviceProfile{
	Name:                     "Chrome Desktop",
	SupportedVideoCodecs:     []string{CodecH264, CodecVP9, CodecAV1},
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecFLAC},
	SupportedContainers:      []string{ContainerMP4, ContainerWebM, ContainerMKV, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 0, // Unlimited
	MaxHeight:                0,
	MaxBitrate:               0,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             true,
}

// FirefoxDesktop profile for Firefox on desktop
var FirefoxDesktop = DeviceProfile{
	Name:                     "Firefox Desktop",
	SupportedVideoCodecs:     []string{CodecH264, CodecVP9, CodecAV1},
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecFLAC},
	SupportedContainers:      []string{ContainerMP4, ContainerWebM, ContainerMKV, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 0,
	MaxHeight:                0,
	MaxBitrate:               0,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             true,
}

// SafariDesktop profile for Safari on macOS
// macOS natively renders HDR/DV via AVFoundation (HEVC Main 10, HLG, Dolby Vision)
var SafariDesktop = DeviceProfile{
	Name:                     "Safari Desktop",
	SupportedVideoCodecs:     []string{CodecH264, CodecH265}, // HEVC on macOS Safari
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecFLAC, CodecAC3, CodecEAC3},
	SupportedContainers:      []string{ContainerMP4, ContainerMOV, ContainerHLS, ContainerMKV},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 0,
	MaxHeight:                0,
	MaxBitrate:               0,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             false, // Safari has limited WebM support
}

// MobileSafari profile for Safari on iOS/iPadOS
// iOS natively decodes AC3/EAC3 via hardware (Apple A-series chips)
// iOS/iPadOS renders HDR/DV natively via AVFoundation on supported devices
var MobileSafari = DeviceProfile{
	Name:                     "Mobile Safari",
	SupportedVideoCodecs:     []string{CodecH264, CodecH265},
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecAC3, CodecEAC3},
	SupportedContainers:      []string{ContainerMP4, ContainerMOV, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 0,
	MaxHeight:                2160,  // 4K on newer devices
	MaxBitrate:               40000, // 40 Mbps for mobile
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             false,
}

// EdgeDesktop profile for Microsoft Edge on desktop
var EdgeDesktop = DeviceProfile{
	Name:                     "Edge Desktop",
	SupportedVideoCodecs:     []string{CodecH264, CodecVP9, CodecAV1, CodecH265}, // Edge supports HEVC on Windows
	SupportedAudioCodecs:     []string{CodecAAC, CodecOpus, CodecMP3, CodecFLAC, CodecAC3, CodecEAC3},
	SupportedContainers:      []string{ContainerMP4, ContainerWebM, ContainerMKV, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 0,
	MaxHeight:                0,
	MaxBitrate:               0,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             true,
}

// GenericBrowser safe fallback profile
var GenericBrowser = DeviceProfile{
	Name:                     "Generic Browser",
	SupportedVideoCodecs:     []string{CodecH264},
	SupportedAudioCodecs:     []string{CodecAAC, CodecMP3},
	SupportedContainers:      []string{ContainerMP4, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 0,
	MaxHeight:                1080,  // Conservative default
	MaxBitrate:               20000, // 20 Mbps
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             false,
}

// SmartTV profile for embedded browsers in smart TVs
var SmartTV = DeviceProfile{
	Name:                     "Smart TV",
	SupportedVideoCodecs:     []string{CodecH264},
	SupportedAudioCodecs:     []string{CodecAAC, CodecAC3},
	SupportedContainers:      []string{ContainerMP4, ContainerHLS},
	SupportedSubtitleFormats: []string{SubtitleVTT},
	MaxWidth:                 1920,
	MaxHeight:                1080,
	MaxBitrate:               15000,
	CanBurnSubtitles:         false,
	SupportsHLS:              true,
	SupportsWebM:             false,
}

// GetBuiltinProfile returns a built-in profile by name
func GetBuiltinProfile(name string) *DeviceProfile {
	switch name {
	case "android_native":
		return &AndroidNative
	case "velox_desktop":
		return &VeloxDesktop
	case "chrome":
		return &ChromeDesktop
	case "firefox":
		return &FirefoxDesktop
	case "safari":
		return &SafariDesktop
	case "mobile_safari":
		return &MobileSafari
	case "edge":
		return &EdgeDesktop
	case "smarttv":
		return &SmartTV
	default:
		return &GenericBrowser
	}
}

// AllBuiltinProfiles returns a map of all available profiles
func AllBuiltinProfiles() map[string]*DeviceProfile {
	return map[string]*DeviceProfile{
		"android_native": &AndroidNative,
		"velox_desktop":  &VeloxDesktop,
		"chrome":         &ChromeDesktop,
		"firefox":        &FirefoxDesktop,
		"safari":         &SafariDesktop,
		"mobile_safari":  &MobileSafari,
		"edge":           &EdgeDesktop,
		"smarttv":        &SmartTV,
		"generic":        &GenericBrowser,
	}
}
