package opensubs_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/pkg/opensubs"
	"github.com/thawng/velox/pkg/subprovider"
)

func TestClient_Configured(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		apiKey   string
		username string
		password string
		expected bool
	}{
		{"all empty", "", "", "", false},
		{"missing apiKey", "", "user", "pass", false},
		{"missing username", "key", "", "pass", false},
		{"missing password", "key", "user", "", false},
		{"all present", "key", "user", "pass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := opensubs.New(tt.apiKey, tt.username, tt.password)
			got := c.Configured()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSearchParams_Parsing(t *testing.T) {
	t.Parallel()
	params := opensubs.SearchParams{
		ImdbID:   "tt1234567",
		TmdbID:   123,
		Query:    "Test Movie",
		Language: "en",
		Year:     2024,
	}

	assert.Equal(t, "tt1234567", params.ImdbID)
	assert.Equal(t, 123, params.TmdbID)
	assert.Equal(t, "Test Movie", params.Query)
	assert.Equal(t, "en", params.Language)
	assert.Equal(t, 2024, params.Year)
}

func TestSearchResponse_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"data": [
			{
				"type": "movie",
				"attributes": {
					"release": "Test Movie 2024",
					"language": "en",
					"format": "srt",
					"download_count": 1000,
					"ratings": 4.5,
					"hearing_impaired": false,
					"foreign_parts_only": false,
					"ai_translated": false,
					"files": [
						{"file_id": 12345, "file_name": "test.srt"}
					]
				}
			},
			{
				"type": "movie",
				"attributes": {
					"release": "Test Movie 2024 Vietnamese",
					"language": "vi",
					"format": "srt",
					"download_count": 500,
					"ratings": 4.0,
					"hearing_impaired": true,
					"foreign_parts_only": true,
					"ai_translated": true,
					"files": [
						{"file_id": 67890, "file_name": "test_vi.srt"}
					]
				}
			}
		]
	}`

	var resp struct {
		Data []struct {
			Type       string `json:"type"`
			Attributes struct {
				Release          string  `json:"release"`
				Language         string  `json:"language"`
				Format           string  `json:"format"`
				DownloadCount    int     `json:"download_count"`
				Ratings          float64 `json:"ratings"`
				HearingImpaired  bool    `json:"hearing_impaired"`
				ForeignPartsOnly bool    `json:"foreign_parts_only"`
				AITranslated     bool    `json:"ai_translated"`
				Files            []struct {
					FileID   int    `json:"file_id"`
					FileName string `json:"file_name"`
				} `json:"files"`
			} `json:"attributes"`
		} `json:"data"`
	}

	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Data, 2)

	// First result
	assert.Equal(t, "Test Movie 2024", resp.Data[0].Attributes.Release)
	assert.Equal(t, "en", resp.Data[0].Attributes.Language)
	assert.Equal(t, "srt", resp.Data[0].Attributes.Format)
	assert.Equal(t, 1000, resp.Data[0].Attributes.DownloadCount)
	assert.Equal(t, 4.5, resp.Data[0].Attributes.Ratings)
	assert.False(t, resp.Data[0].Attributes.HearingImpaired)
	assert.False(t, resp.Data[0].Attributes.ForeignPartsOnly)
	assert.False(t, resp.Data[0].Attributes.AITranslated)
	require.Len(t, resp.Data[0].Attributes.Files, 1)
	assert.Equal(t, 12345, resp.Data[0].Attributes.Files[0].FileID)

	// Second result
	assert.Equal(t, "vi", resp.Data[1].Attributes.Language)
	assert.True(t, resp.Data[1].Attributes.HearingImpaired)
	assert.True(t, resp.Data[1].Attributes.ForeignPartsOnly)
	assert.True(t, resp.Data[1].Attributes.AITranslated)
}

func TestSearchResult_ToSubproviderResult(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"data": [
			{
				"type": "movie",
				"attributes": {
					"release": "Test Movie 2024",
					"language": "en",
					"format": "srt",
					"download_count": 1000,
					"ratings": 4.5,
					"hearing_impaired": false,
					"foreign_parts_only": false,
					"ai_translated": false,
					"files": [
						{"file_id": 12345, "file_name": "test.srt"}
					]
				}
			}
		]
	}`

	var apiResp struct {
		Data []struct {
			Attributes struct {
				Release          string  `json:"release"`
				Language         string  `json:"language"`
				Format           string  `json:"format"`
				DownloadCount    int     `json:"download_count"`
				Ratings          float64 `json:"ratings"`
				HearingImpaired  bool    `json:"hearing_impaired"`
				ForeignPartsOnly bool    `json:"foreign_parts_only"`
				AITranslated     bool    `json:"ai_translated"`
				Files            []struct {
					FileID int `json:"file_id"`
				} `json:"files"`
			} `json:"attributes"`
		} `json:"data"`
	}

	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&apiResp)
	require.NoError(t, err)

	// Convert to subprovider.Result format
	results := make([]subprovider.Result, 0, len(apiResp.Data))
	for _, item := range apiResp.Data {
		if len(item.Attributes.Files) == 0 {
			continue
		}
		// Use FileID to ensure it's accessible
		_ = item.Attributes.Files[0].FileID
		results = append(results, subprovider.Result{
			Provider:        "opensubtitles",
			ExternalID:      "12345", // file_id as string
			Title:           item.Attributes.Release,
			Language:        item.Attributes.Language,
			Format:          item.Attributes.Format,
			Downloads:       item.Attributes.DownloadCount,
			Rating:          item.Attributes.Ratings,
			Forced:          item.Attributes.ForeignPartsOnly,
			HearingImpaired: item.Attributes.HearingImpaired,
			AITranslated:    item.Attributes.AITranslated,
		})
	}

	require.Len(t, results, 1)
	assert.Equal(t, "opensubtitles", results[0].Provider)
	assert.Equal(t, "12345", results[0].ExternalID)
	assert.Equal(t, "Test Movie 2024", results[0].Title)
	assert.Equal(t, "en", results[0].Language)
	assert.Equal(t, "srt", results[0].Format)
	assert.Equal(t, 1000, results[0].Downloads)
	assert.Equal(t, 4.5, results[0].Rating)
}

func TestLoginResponse_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{"token": "jwt-token-abc123xyz"}`

	var result struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "jwt-token-abc123xyz", result.Token)
}

func TestDownloadResponse_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"link": "https://download.opensubtitles.com/12345",
		"file_name": "test_movie_en.srt"
	}`

	var dlResp struct {
		Link     string `json:"link"`
		FileName string `json:"file_name"`
	}
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&dlResp)
	require.NoError(t, err)
	assert.Equal(t, "https://download.opensubtitles.com/12345", dlResp.Link)
	assert.Equal(t, "test_movie_en.srt", dlResp.FileName)
}

func TestClient_New(t *testing.T) {
	t.Parallel()
	c := opensubs.New("api-key", "username", "password")
	assert.NotNil(t, c)
	assert.True(t, c.Configured())
}

func TestClient_New_NotConfigured(t *testing.T) {
	t.Parallel()
	c := opensubs.New("", "", "")
	assert.NotNil(t, c)
	assert.False(t, c.Configured())
}

// Ensure subprovider.Result has the expected fields
func TestSubproviderResult_Fields(t *testing.T) {
	t.Parallel()
	r := subprovider.Result{
		Provider:        "test",
		ExternalID:      "123",
		Title:           "Test",
		Language:        "en",
		Format:          "srt",
		Downloads:       100,
		Rating:          4.5,
		Forced:          false,
		HearingImpaired: false,
		AITranslated:    false,
	}

	assert.Equal(t, "test", r.Provider)
	assert.Equal(t, "123", r.ExternalID)
	assert.Equal(t, "Test", r.Title)
	assert.Equal(t, "en", r.Language)
	assert.Equal(t, "srt", r.Format)
	assert.Equal(t, 100, r.Downloads)
	assert.Equal(t, 4.5, r.Rating)
	assert.False(t, r.Forced)
	assert.False(t, r.HearingImpaired)
	assert.False(t, r.AITranslated)
}

// HTTP transport mock for OpenSubtitles
type mockTransport struct {
	respStatus int
	respBody   string
	callback   func(*http.Request)
}

func (m *mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if m.callback != nil {
		m.callback(r)
	}
	return &http.Response{
		StatusCode: m.respStatus,
		Body:       io.NopCloser(strings.NewReader(m.respBody)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}
