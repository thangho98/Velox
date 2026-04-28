package omdb_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/pkg/omdb"
)

func TestClient_New(t *testing.T) {
	t.Parallel()
	c := omdb.New("test-api-key")
	assert.NotNil(t, c)
}

func TestResponse_IMDbRatingFloat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rating   string
		expected float64
	}{
		{"8.5", 8.5},
		{"7.0", 7.0},
		{"invalid", 0},
		{"", 0},
		{"10.0", 10.0},
		{"0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.rating, func(t *testing.T) {
			r := &omdb.Response{IMDbRating: tt.rating}
			got := r.IMDbRatingFloat()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResponse_RottenTomatoesScore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ratings  []omdb.Rating
		expected int
	}{
		{
			name:     "RT 92%",
			ratings:  []omdb.Rating{{Source: "Rotten Tomatoes", Value: "92%"}},
			expected: 92,
		},
		{
			name:     "RT 100%",
			ratings:  []omdb.Rating{{Source: "Rotten Tomatoes", Value: "100%"}},
			expected: 100,
		},
		{
			name:     "no RT",
			ratings:  []omdb.Rating{{Source: "Metacritic", Value: "88/100"}},
			expected: 0,
		},
		{
			name:     "empty",
			ratings:  []omdb.Rating{},
			expected: 0,
		},
		{
			name:     "invalid format",
			ratings:  []omdb.Rating{{Source: "Rotten Tomatoes", Value: "invalid"}},
			expected: 0,
		},
		{
			name:     "no percent sign",
			ratings:  []omdb.Rating{{Source: "Rotten Tomatoes", Value: "92"}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &omdb.Response{Ratings: tt.ratings}
			got := r.RottenTomatoesScore()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResponse_MetascoreInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		metascore string
		expected  int
	}{
		{"88", 88},
		{"100", 100},
		{"0", 0},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.metascore, func(t *testing.T) {
			r := &omdb.Response{Metascore: tt.metascore}
			got := r.MetascoreInt()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResponse_JSONParsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"Title": "Test Movie",
		"Year": "2024",
		"Rated": "PG-13",
		"Released": "15 Jan 2024",
		"Runtime": "120 min",
		"Genre": "Action, Adventure",
		"Director": "Test Director",
		"Writer": "Test Writer",
		"Actors": "Actor One, Actor Two",
		"Plot": "A test movie plot",
		"Language": "English",
		"Country": "USA",
		"Awards": "2 wins",
		"Poster": "https://example.com/poster.jpg",
		"Ratings": [
			{"Source": "Internet Movie Database", "Value": "8.5/10"}
		],
		"Metascore": "88",
		"imdbRating": "8.5",
		"imdbVotes": "1,000,000",
		"imdbID": "tt1234567",
		"Type": "movie",
		"BoxOffice": "$100,000,000",
		"Response": "True"
	}`

	var resp omdb.Response
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "Test Movie", resp.Title)
	assert.Equal(t, "2024", resp.Year)
	assert.Equal(t, "PG-13", resp.Rated)
	assert.Equal(t, "15 Jan 2024", resp.Released)
	assert.Equal(t, "120 min", resp.Runtime)
	assert.Equal(t, "Action, Adventure", resp.Genre)
	assert.Equal(t, "Test Director", resp.Director)
	assert.Equal(t, "Test Writer", resp.Writer)
	assert.Equal(t, "Actor One, Actor Two", resp.Actors)
	assert.Equal(t, "A test movie plot", resp.Plot)
	assert.Equal(t, "English", resp.Language)
	assert.Equal(t, "USA", resp.Country)
	assert.Equal(t, "2 wins", resp.Awards)
	assert.Equal(t, "https://example.com/poster.jpg", resp.Poster)
	assert.Len(t, resp.Ratings, 1)
	assert.Equal(t, "Internet Movie Database", resp.Ratings[0].Source)
	assert.Equal(t, "8.5/10", resp.Ratings[0].Value)
	assert.Equal(t, "88", resp.Metascore)
	assert.Equal(t, "8.5", resp.IMDbRating)
	assert.Equal(t, "1,000,000", resp.IMDbVotes)
	assert.Equal(t, "tt1234567", resp.IMDbID)
	assert.Equal(t, "movie", resp.Type)
	assert.Equal(t, "$100,000,000", resp.BoxOffice)
	assert.Equal(t, "True", resp.Response)
}

func TestResponse_ErrorParsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"Response": "False",
		"Error": "Movie not found!"
	}`

	var resp omdb.Response
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "False", resp.Response)
	assert.Equal(t, "Movie not found!", resp.Error)
}

func TestResponse_JSONParsing_Series(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"Title": "Test Series",
		"Year": "2024",
		"Type": "series",
		"Response": "True"
	}`

	var resp omdb.Response
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Test Series", resp.Title)
	assert.Equal(t, "series", resp.Type)
}

func TestRating_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{"Source": "Rotten Tomatoes", "Value": "93%"}`

	var rating omdb.Rating
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&rating)
	require.NoError(t, err)
	assert.Equal(t, "Rotten Tomatoes", rating.Source)
	assert.Equal(t, "93%", rating.Value)
}
