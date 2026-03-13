package subprovider

// Result represents a subtitle search result from any external provider.
type Result struct {
	Provider        string  `json:"provider"`         // "opensubtitles" | "podnapisi"
	ExternalID      string  `json:"external_id"`      // provider's file/subtitle ID
	Title           string  `json:"title"`            // release name
	Language        string  `json:"language"`         // ISO 639-1 ("en", "vi")
	Format          string  `json:"format"`           // "srt", "sub", "ass"
	Downloads       int     `json:"downloads"`        // download count
	Rating          float64 `json:"rating"`           // provider-specific rating
	Forced          bool    `json:"forced"`           // forced subtitle flag
	HearingImpaired bool    `json:"hearing_impaired"` // SDH/CC flag
	AITranslated    bool    `json:"ai_translated"`    // machine-translated flag
}
