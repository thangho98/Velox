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
