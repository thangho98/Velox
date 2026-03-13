package fanart

import "errors"

// ErrNotFound is returned when fanart.tv has no images for the given ID.
var ErrNotFound = errors.New("fanart: not found")

// Image represents a single artwork entry from fanart.tv.
type Image struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Lang  string `json:"lang"`
	Likes string `json:"likes"`
}

// MovieImages contains all artwork types for a movie.
type MovieImages struct {
	Name        string  `json:"name"`
	TmdbID      string  `json:"tmdb_id"`
	IMDbID      string  `json:"imdb_id"`
	HDClearArt  []Image `json:"hdmovieclearart"`
	HDClearLogo []Image `json:"hdmovielogo"`
	MoviePoster []Image `json:"movieposter"`
	MovieBanner []Image `json:"moviebanner"`
	MovieDisc   []Image `json:"moviedisc"`
	MovieThumb  []Image `json:"moviethumb"`
	MovieArt    []Image `json:"movieart"`
	MovieBG     []Image `json:"moviebackground"`
	MovieLogo   []Image `json:"movielogo"`
}

// ShowImages contains all artwork types for a TV show.
type ShowImages struct {
	Name         string  `json:"name"`
	TvdbID       string  `json:"thetvdb_id"`
	HDClearArt   []Image `json:"hdclearart"`
	HDClearLogo  []Image `json:"hdtvlogo"`
	TVPoster     []Image `json:"tvposter"`
	TVBanner     []Image `json:"tvbanner"`
	TVThumb      []Image `json:"tvthumb"`
	ShowBG       []Image `json:"showbackground"`
	SeasonPoster []Image `json:"seasonposter"`
	SeasonBanner []Image `json:"seasonbanner"`
	SeasonThumb  []Image `json:"seasonthumb"`
	CharacterArt []Image `json:"characterart"`
	ClearLogo    []Image `json:"clearlogo"`
	ClearArt     []Image `json:"clearart"`
}

// BestImage returns the first image URL from the slice, preferring English,
// or empty string if none available.
func BestImage(images []Image) string {
	if len(images) == 0 {
		return ""
	}
	// Prefer English
	for _, img := range images {
		if img.Lang == "en" {
			return img.URL
		}
	}
	// Fallback: first image with no language tag or any language
	for _, img := range images {
		if img.Lang == "" || img.Lang == "00" {
			return img.URL
		}
	}
	return images[0].URL
}
