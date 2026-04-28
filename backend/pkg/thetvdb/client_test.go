package thetvdb_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/pkg/thetvdb"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClient_Search(t *testing.T) {
	t.Parallel()
	// Test SearchResult parsing directly via JSON
	data := `{
		"status": "success",
		"data": [
			{
				"objectID":"series-123",
				"tvdb_id":"123",
				"name":"Test Show",
				"type":"series",
				"year":"2024",
				"image_url":"https://example.com/poster.jpg",
				"overview":"A test show",
				"status":"Ended",
				"slug":"test-show",
				"first_air_time":"2024-01-15",
				"primary_language":"eng",
				"aliases":["Test"]
			}
		]
	}`

	var resp struct {
		Status string                 `json:"status"`
		Data   []thetvdb.SearchResult `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "series-123", resp.Data[0].ObjectID)
	assert.Equal(t, "123", resp.Data[0].TVDBID)
	assert.Equal(t, "Test Show", resp.Data[0].Name)
	assert.Equal(t, "series", resp.Data[0].Type)
	assert.Equal(t, "2024", resp.Data[0].Year)
	assert.Equal(t, "https://example.com/poster.jpg", resp.Data[0].ImageURL)
	assert.Contains(t, resp.Data[0].Overview, "A test show")
	assert.Equal(t, "Ended", resp.Data[0].Status)
	assert.Equal(t, "test-show", resp.Data[0].Slug)
	assert.Equal(t, "2024-01-15", resp.Data[0].FirstAirTime)
	assert.Equal(t, "eng", resp.Data[0].PrimaryLanguage)
	assert.Len(t, resp.Data[0].Aliases, 1)
}

func TestClient_GetSeries(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"id": 81189,
			"name": "The Simpsons",
			"slug": "the-simpsons",
			"image": "/banners/posters/81189.jpg",
			"firstAired": "1989-12-17",
			"lastAired": "2024-05-22",
			"averageRuntime": 22,
			"originalCountry": "us",
			"originalLanguage": "eng",
			"overview": "The Simpsons is an American animated sitcom.",
			"year": "1989"
		}
	}`

	var resp struct {
		Status string             `json:"status"`
		Data   thetvdb.SeriesBase `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 81189, resp.Data.ID)
	assert.Equal(t, "The Simpsons", resp.Data.Name)
	assert.Equal(t, 22, resp.Data.AverageRuntime)
}

func TestClient_GetSeriesEpisodes(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"series": {"id": 81189, "name": "The Simpsons"},
			"episodes": [
				{
					"id": 553,
					"seriesId": 81189,
					"name": "Simpsons Roasting on an Open Fire",
					"aired": "1989-12-17",
					"runtime": 22,
					"number": 1,
					"seasonNumber": 1,
					"absoluteNumber": 1
				}
			]
		}
	}`

	var resp struct {
		Status string                         `json:"status"`
		Data   thetvdb.SeriesEpisodesResponse `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 81189, resp.Data.Series.ID)
	require.Len(t, resp.Data.Episodes, 1)
	assert.Equal(t, "Simpsons Roasting on an Open Fire", resp.Data.Episodes[0].Name)
	assert.Equal(t, 1, resp.Data.Episodes[0].Number)
}

func TestClient_GetEpisode(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"id": 553,
			"seriesId": 81189,
			"name": "Simpsons Roasting on an Open Fire",
			"aired": "1989-12-17",
			"runtime": 22,
			"overview": "The family is forced to spend Christmas at the mall.",
			"image": "/banners/episodes/81189-553.jpg",
			"number": 1,
			"seasonNumber": 1,
			"absoluteNumber": 1,
			"year": "1989"
		}
	}`

	var resp struct {
		Status string              `json:"status"`
		Data   thetvdb.EpisodeBase `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 553, resp.Data.ID)
	assert.Equal(t, 22, resp.Data.Runtime)
}

func TestClient_GetEpisodeExtended(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"id": 553,
			"seriesId": 81189,
			"name": "Simpsons Roasting on an Open Fire",
			"aired": "1989-12-17",
			"runtime": 22,
			"number": 1,
			"seasonNumber": 1,
			"characters": [
				{
					"id": 1,
					"peopleId": 100,
					"personName": "Dan Castellaneta",
					"peopleType": "Actor",
					"name": "Homer Simpson",
					"image": "/banners/characters/100.jpg",
					"sort": 1,
					"isFeatured": true
				}
			],
			"contentRatings": [
				{"id": 1, "name": "TV-G", "country": "us"}
			],
			"remoteIds": [
				{"id": "tt0096697", "type": 2, "sourceName": "IMDB"}
			]
		}
	}`

	var resp struct {
		Status string                  `json:"status"`
		Data   thetvdb.EpisodeExtended `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data.Characters, 1)
	assert.Equal(t, "Homer Simpson", resp.Data.Characters[0].Name)
	assert.Len(t, resp.Data.RemoteIDs, 1)
}

func TestClient_GetSeason(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"id": 100,
			"seriesId": 81189,
			"number": 1,
			"name": "Season 1",
			"image": "/banners/seasons/100.jpg",
			"imageType": 2,
			"lastUpdated": "2024-01-01",
			"type": {"id": 1, "name": "Official", "type": "official"}
		}
	}`

	var resp struct {
		Status string             `json:"status"`
		Data   thetvdb.SeasonBase `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 100, resp.Data.ID)
	assert.Equal(t, 1, resp.Data.Number)
	require.NotNil(t, resp.Data.SeasonType)
	assert.Equal(t, "Official", resp.Data.SeasonType.Name)
}

func TestClient_GetMovie(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"id": 100,
			"name": "Test Movie",
			"slug": "test-movie",
			"image": "/banners/movies/100.jpg",
			"score": 85,
			"runtime": 120,
			"year": "2024",
			"overview": "A test movie",
			"status": {"id": 1, "name": "Released", "recordType": "movie", "keepUpdated": false}
		}
	}`

	var resp struct {
		Status string            `json:"status"`
		Data   thetvdb.MovieBase `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 100, resp.Data.ID)
	assert.Equal(t, 120, resp.Data.Runtime)
}

func TestClient_GetSeriesTranslation(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"name": "Doctor Who",
			"overview": "The long-running British science fiction series.",
			"language": "eng",
			"isPrimary": true
		}
	}`

	var resp struct {
		Status string                    `json:"status"`
		Data   thetvdb.TranslationRecord `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Doctor Who", resp.Data.Name)
	assert.True(t, resp.Data.IsPrimary)
}

func TestIMDbID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		remoteIDs []thetvdb.RemoteID
		expected  string
	}{
		{
			name: "IMDB by type",
			remoteIDs: []thetvdb.RemoteID{
				{ID: "tt1234567", Type: 2, SourceName: "IMDB"},
				{ID: "tvdb-456", Type: 1, SourceName: "TVDB"},
			},
			expected: "tt1234567",
		},
		{
			name: "IMDB by source name",
			remoteIDs: []thetvdb.RemoteID{
				{ID: "tt9999999", Type: 99, SourceName: "IMDB"},
			},
			expected: "tt9999999",
		},
		{
			name: "no IMDB",
			remoteIDs: []thetvdb.RemoteID{
				{ID: "tvdb-456", Type: 1, SourceName: "TVDB"},
			},
			expected: "",
		},
		{
			name:      "empty list",
			remoteIDs: []thetvdb.RemoteID{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thetvdb.IMDbID(tt.remoteIDs)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLoginResponse_Parsing(t *testing.T) {
	t.Parallel()
	// Login response is just the token directly, not wrapped in apiResponse
	data := `{"token": "jwt-token-abc123xyz"}`

	var loginResp thetvdb.LoginResponse
	err := json.NewDecoder(strings.NewReader(data)).Decode(&loginResp)
	require.NoError(t, err)
	assert.Equal(t, "jwt-token-abc123xyz", loginResp.Token)
}

func TestClient_GetSeriesArtworks(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": {
			"id": 81189,
			"name": "The Simpsons",
			"artworks": [
				{
					"id": 1,
					"image": "/banners/artwork/1.jpg",
					"thumbnail": "/banners/artwork/thumb/1.jpg",
					"language": "en",
					"type": 2,
					"score": 100,
					"width": 1920,
					"height": 1080
				},
				{
					"id": 2,
					"image": "/banners/artwork/2.jpg",
					"thumbnail": "/banners/artwork/thumb/2.jpg",
					"language": "",
					"type": 1,
					"score": 90
				}
			]
		}
	}`

	var resp struct {
		Status string             `json:"status"`
		Data   thetvdb.SeriesBase `json:"data"`
	}
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Data.Artworks, 2)
	assert.Equal(t, "en", resp.Data.Artworks[0].Language)
	assert.Equal(t, 2, resp.Data.Artworks[0].Type)
}

func TestStatusRecord_Parsing(t *testing.T) {
	t.Parallel()
	data := `{"id": 1, "name": "Continuing", "recordType": "series", "keepUpdated": true}`

	var status thetvdb.StatusRecord
	err := json.Unmarshal([]byte(data), &status)
	require.NoError(t, err)
	assert.Equal(t, "Continuing", status.Name)
	assert.True(t, status.KeepUpdated)
}

func TestGenreRecord_Parsing(t *testing.T) {
	t.Parallel()
	data := `{"id": 2, "name": "Animation", "slug": "animation"}`

	var genre thetvdb.GenreRecord
	err := json.Unmarshal([]byte(data), &genre)
	require.NoError(t, err)
	assert.Equal(t, "Animation", genre.Name)
	assert.Equal(t, "animation", genre.Slug)
}

func TestType_Parsing(t *testing.T) {
	t.Parallel()
	data := `{"id": 1, "name": "Official", "type": "official"}`

	var seasonType thetvdb.Type
	err := json.Unmarshal([]byte(data), &seasonType)
	require.NoError(t, err)
	assert.Equal(t, "Official", seasonType.Name)
	assert.Equal(t, "official", seasonType.Type)
}

// Ensure the Client struct can be created
func TestClient_Creation(t *testing.T) {
	t.Parallel()
	c := thetvdb.New("test-api-key")
	assert.NotNil(t, c)
}

// apiResponse is the unexported response wrapper used in the package.
// We test the JSON parsing by using an equivalent local struct.
type apiResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

func TestAPIResponseWrapper(t *testing.T) {
	t.Parallel()
	type testData struct {
		Name string `json:"name"`
	}

	data := `{"status": "success", "data": {"name": "test"}}`

	var resp apiResponse[testData]
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "test", resp.Data.Name)
}

func TestAPIResponseWrapper_List(t *testing.T) {
	t.Parallel()
	data := `{
		"status": "success",
		"data": [
			{"name": "first"},
			{"name": "second"}
		]
	}`

	var resp apiResponse[[]struct{ Name string }]
	err := json.NewDecoder(strings.NewReader(data)).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "first", resp.Data[0].Name)
}
