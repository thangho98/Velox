package model

// AppSetting represents a single key-value setting row.
type AppSetting struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

// Known setting keys for OpenSubtitles integration.
const (
	SettingOpenSubsAPIKey   = "opensubs_api_key"
	SettingOpenSubsUsername = "opensubs_username"
	SettingOpenSubsPassword = "opensubs_password"
)

// Known setting keys for TMDb integration.
const (
	SettingTMDbAPIKey = "tmdb_api_key"
)

// Known setting keys for OMDb integration.
const (
	SettingOMDbAPIKey = "omdb_api_key"
)

// Known setting keys for TheTVDB integration.
const (
	SettingTVDBAPIKey = "tvdb_api_key"
)

// Known setting keys for fanart.tv integration.
const (
	SettingFanartAPIKey = "fanart_api_key"
)

// Known setting keys for Subdl integration.
const (
	SettingSubdlAPIKey = "subdl_api_key"
)

// Known setting keys for subtitle auto-download.
const (
	// SettingAutoSubLanguages is a comma-separated list of language codes to auto-download.
	// Example: "en,vi". Empty string disables auto-download.
	SettingAutoSubLanguages = "auto_sub_languages"
)

// Known setting keys for subtitle translation.
const (
	SettingDeepLAPIKey = "deepl_api_key"
)

// Known setting keys for playback policy.
const (
	// SettingPlaybackMode controls server-wide playback behavior.
	// Values: "auto" (default, decide based on client), "direct_play" (force direct play, never transcode)
	SettingPlaybackMode = "playback_mode"
)
