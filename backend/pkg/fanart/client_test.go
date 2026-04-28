package fanart_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/pkg/fanart"
)

func TestClient_New(t *testing.T) {
	t.Parallel()
	c := fanart.New("test-api-key")
	assert.NotNil(t, c)
}

func TestBestImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		images   []fanart.Image
		expected string
	}{
		{
			name:     "empty",
			images:   []fanart.Image{},
			expected: "",
		},
		{
			name: "prefer english",
			images: []fanart.Image{
				{URL: "https://example.com/fr.jpg", Lang: "fr"},
				{URL: "https://example.com/en.jpg", Lang: "en"},
				{URL: "https://example.com/de.jpg", Lang: "de"},
			},
			expected: "https://example.com/en.jpg",
		},
		{
			name: "fallback to no lang",
			images: []fanart.Image{
				{URL: "https://example.com/fr.jpg", Lang: "fr"},
				{URL: "https://example.com/none.jpg", Lang: ""},
			},
			expected: "https://example.com/none.jpg",
		},
		{
			name: "fallback to 00",
			images: []fanart.Image{
				{URL: "https://example.com/fr.jpg", Lang: "fr"},
				{URL: "https://example.com/00.jpg", Lang: "00"},
			},
			expected: "https://example.com/00.jpg",
		},
		{
			name: "first if no english/no lang",
			images: []fanart.Image{
				{URL: "https://example.com/fr.jpg", Lang: "fr"},
				{URL: "https://example.com/de.jpg", Lang: "de"},
			},
			expected: "https://example.com/fr.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fanart.BestImage(tt.images)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, fanart.ErrNotFound)
	assert.True(t, errors.Is(fanart.ErrNotFound, fanart.ErrNotFound))
}

func TestImage_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": "12345",
		"url": "https://assets.fanart.tv/uploads/premium/12345.jpg",
		"lang": "en",
		"likes": "100"
	}`

	var img fanart.Image
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&img)
	require.NoError(t, err)
	assert.Equal(t, "12345", img.ID)
	assert.Equal(t, "https://assets.fanart.tv/uploads/premium/12345.jpg", img.URL)
	assert.Equal(t, "en", img.Lang)
	assert.Equal(t, "100", img.Likes)
}

func TestMovieImages_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"name": "Test Movie",
		"tmdb_id": "12345",
		"imdb_id": "tt1234567",
		"hdmovieclearart": [
			{"id": "1", "url": "https://example.com/clearart.png", "lang": "en", "likes": "50"}
		],
		"hdmovielogo": [
			{"id": "2", "url": "https://example.com/logo.png", "lang": "en", "likes": "40"}
		],
		"movieposter": [
			{"id": "3", "url": "https://example.com/poster.jpg", "lang": "en", "likes": "100"}
		],
		"moviebanner": [],
		"moviedisc": [],
		"moviethumb": [],
		"movieart": [],
		"moviebackground": [
			{"id": "4", "url": "https://example.com/bg.jpg", "lang": "", "likes": "30"}
		],
		"movielogo": []
	}`

	var images fanart.MovieImages
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&images)
	require.NoError(t, err)

	assert.Equal(t, "Test Movie", images.Name)
	assert.Equal(t, "12345", images.TmdbID)
	assert.Equal(t, "tt1234567", images.IMDbID)
	require.Len(t, images.HDClearArt, 1)
	assert.Equal(t, "https://example.com/clearart.png", images.HDClearArt[0].URL)
	require.Len(t, images.MoviePoster, 1)
	assert.Equal(t, "https://example.com/poster.jpg", images.MoviePoster[0].URL)
	require.Len(t, images.MovieBG, 1)
	assert.Equal(t, "", images.MovieBG[0].Lang)
}

func TestShowImages_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"name": "Test Show",
		"thetvdb_id": "123456",
		"hdclearart": [
			{"id": "1", "url": "https://example.com/clearart.png", "lang": "en", "likes": "50"}
		],
		"hdtvlogo": [
			{"id": "2", "url": "https://example.com/logo.png", "lang": "en", "likes": "40"}
		],
		"tvposter": [
			{"id": "3", "url": "https://example.com/poster.jpg", "lang": "en", "likes": "100"}
		],
		"tvbanner": [],
		"tvthumb": [],
		"showbackground": [],
		"seasonposter": [],
		"seasonbanner": [],
		"seasonthumb": [],
		"characterart": [],
		"clearlogo": [],
		"clearart": []
	}`

	var images fanart.ShowImages
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&images)
	require.NoError(t, err)

	assert.Equal(t, "Test Show", images.Name)
	assert.Equal(t, "123456", images.TvdbID)
	require.Len(t, images.HDClearArt, 1)
	require.Len(t, images.TVPoster, 1)
	assert.Equal(t, "https://example.com/poster.jpg", images.TVPoster[0].URL)
}

func TestClient_GetMovieImages(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"name": "Test Movie",
		"tmdb_id": "12345",
		"movieposter": [
			{"id": "1", "url": "https://example.com/poster.jpg", "lang": "en", "likes": "100"}
		]
	}`

	var images fanart.MovieImages
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&images)
	require.NoError(t, err)
	assert.Equal(t, "Test Movie", images.Name)
	assert.Equal(t, "12345", images.TmdbID)
}

func TestClient_GetShowImages(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"name": "Test Show",
		"thetvdb_id": "123456",
		"tvposter": [
			{"id": "1", "url": "https://example.com/poster.jpg", "lang": "en", "likes": "100"}
		]
	}`

	var images fanart.ShowImages
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&images)
	require.NoError(t, err)
	assert.Equal(t, "Test Show", images.Name)
	assert.Equal(t, "123456", images.TvdbID)
}

func TestBestImage_EmptySlice(t *testing.T) {
	t.Parallel()
	images := []fanart.Image{}
	got := fanart.BestImage(images)
	assert.Equal(t, "", got)
}

func TestBestImage_SingleEnglish(t *testing.T) {
	t.Parallel()
	images := []fanart.Image{
		{URL: "https://example.com/en.jpg", Lang: "en"},
	}
	got := fanart.BestImage(images)
	assert.Equal(t, "https://example.com/en.jpg", got)
}

func TestBestImage_SingleNoLang(t *testing.T) {
	t.Parallel()
	images := []fanart.Image{
		{URL: "https://example.com/no-lang.jpg", Lang: ""},
	}
	got := fanart.BestImage(images)
	assert.Equal(t, "https://example.com/no-lang.jpg", got)
}

// roundTripFunc implements http.RoundTripper
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

var _ http.RoundTripper = roundTripFunc(nil)
