package tmdb_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/pkg/tmdb"
)

func TestGetYear(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected int
	}{
		{"2024-01-15", 2024},
		{"", 0},
		{"invalid", 0},
		{"2024-12-31", 2024},
		{"1999-06-01", 1999},
	}

	for _, tt := range tests {
		got := tmdb.GetYear(tt.input)
		assert.Equal(t, tt.expected, got, "GetYear(%q)", tt.input)
	}
}

func TestMovieSearchResults_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"page": 1,
		"results": [
			{
				"id": 123,
				"title": "Test Movie",
				"original_title": "Test Movie Original",
				"overview": "A test movie",
				"release_date": "2024-01-15",
				"poster_path": "/poster.jpg",
				"backdrop_path": "/backdrop.jpg",
				"vote_average": 8.5,
				"vote_count": 1000,
				"popularity": 100.5,
				"adult": false,
				"video": false,
				"genre_ids": [1, 2, 3]
			}
		],
		"total_results": 1,
		"total_pages": 1
	}`

	var results tmdb.MovieSearchResults
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&results)
	require.NoError(t, err)
	assert.Equal(t, 1, results.Page)
	require.Len(t, results.Results, 1)
	r := results.Results[0]
	assert.Equal(t, 123, r.ID)
	assert.Equal(t, "Test Movie", r.Title)
	assert.Equal(t, "Test Movie Original", r.OriginalTitle)
	assert.Equal(t, "A test movie", r.Overview)
	assert.Equal(t, "2024-01-15", r.ReleaseDate)
	assert.Equal(t, "/poster.jpg", r.PosterPath)
	assert.Equal(t, "/backdrop.jpg", r.BackdropPath)
	assert.Equal(t, 8.5, r.VoteAverage)
	assert.Equal(t, 1000, r.VoteCount)
	assert.Equal(t, 100.5, r.Popularity)
	assert.False(t, r.Adult)
	assert.False(t, r.Video)
	assert.Equal(t, []int{1, 2, 3}, r.GenreIDs)
	assert.Equal(t, 1, results.TotalResults)
	assert.Equal(t, 1, results.TotalPages)
}

func TestTVSearchResults_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"page": 1,
		"results": [
			{
				"id": 456,
				"name": "Test Show",
				"original_name": "Test Show Original",
				"overview": "A test TV show",
				"first_air_date": "2024-01-15",
				"poster_path": "/poster.jpg",
				"backdrop_path": "/backdrop.jpg",
				"vote_average": 9.0,
				"vote_count": 2000,
				"popularity": 200.0,
				"genre_ids": [2],
				"origin_country": ["US"],
				"original_language": "en"
			}
		],
		"total_results": 1,
		"total_pages": 5
	}`

	var results tmdb.TVSearchResults
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&results)
	require.NoError(t, err)
	require.Len(t, results.Results, 1)
	r := results.Results[0]
	assert.Equal(t, 456, r.ID)
	assert.Equal(t, "Test Show", r.Name)
	assert.Equal(t, "Test Show Original", r.OriginalName)
	assert.Equal(t, "en", r.OriginalLanguage)
	assert.Equal(t, []string{"US"}, r.OriginCountry)
	assert.Equal(t, 5, results.TotalPages)
}

func TestMovieDetails_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 123,
		"title": "Test Movie",
		"imdb_id": "tt1234567",
		"budget": 100000000,
		"revenue": 300000000,
		"runtime": 120,
		"status": "Released",
		"tagline": "A tagline",
		"homepage": "https://example.com",
		"adult": false,
		"video": false,
		"genres": [{"id": 1, "name": "Action"}],
		"production_companies": [{"id": 1, "name": "Studio"}],
		"production_countries": [{"iso_3166_1": "US", "name": "United States"}],
		"spoken_languages": [{"iso_639_1": "en", "name": "English", "english_name": "English"}]
	}`

	var details tmdb.MovieDetails
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&details)
	require.NoError(t, err)
	assert.Equal(t, 123, details.ID)
	assert.Equal(t, "tt1234567", details.IMDbID)
	assert.Equal(t, int64(100000000), details.Budget)
	assert.Equal(t, int64(300000000), details.Revenue)
	assert.Equal(t, 120, details.Runtime)
	assert.Equal(t, "Released", details.Status)
	assert.Equal(t, "A tagline", details.Tagline)
	assert.Equal(t, "https://example.com", details.Homepage)
	assert.False(t, details.Adult)
	assert.False(t, details.Video)
	require.Len(t, details.Genres, 1)
	assert.Equal(t, "Action", details.Genres[0].Name)
}

func TestTVDetails_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 456,
		"name": "Test Show",
		"number_of_seasons": 3,
		"number_of_episodes": 60,
		"genres": [{"id": 2, "name": "Drama"}],
		"external_ids": {"imdb_id": "tt654321"},
		"seasons": [{"id": 100, "season_number": 1}],
		"networks": [{"id": 1, "name": "NBC"}]
	}`

	var details tmdb.TVDetails
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&details)
	require.NoError(t, err)
	assert.Equal(t, 456, details.ID)
	assert.Equal(t, "Test Show", details.Name)
	assert.Equal(t, 3, details.NumberOfSeasons)
	assert.Equal(t, 60, details.NumberOfEpisodes)
	require.NotNil(t, details.ExternalIDs)
	assert.Equal(t, "tt654321", details.ExternalIDs.IMDbID)
	require.Len(t, details.Seasons, 1)
	assert.Equal(t, 1, details.Seasons[0].SeasonNumber)
}

func TestVideoList_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"results": [
			{"id": "v1", "key": "abc123", "name": "Official Trailer", "site": "YouTube", "type": "Trailer", "official": true, "size": 1080},
			{"id": "v2", "key": "xyz789", "name": "Teaser", "site": "YouTube", "type": "Teaser", "official": false, "size": 720}
		]
	}`

	var videos tmdb.VideoList
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&videos)
	require.NoError(t, err)
	require.Len(t, videos.Results, 2)
	assert.Equal(t, "abc123", videos.Results[0].Key)
	assert.Equal(t, "Official Trailer", videos.Results[0].Name)
	assert.Equal(t, "YouTube", videos.Results[0].Site)
	assert.Equal(t, "Trailer", videos.Results[0].Type)
	assert.True(t, videos.Results[0].Official)
	assert.Equal(t, 1080, videos.Results[0].Size)
}

func TestCredits_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"cast": [
			{"id": 1, "name": "Actor One", "character": "Hero", "order": 1, "profile_path": "/profile1.jpg"}
		],
		"crew": [
			{"id": 2, "name": "Director One", "job": "Director", "department": "Directing", "profile_path": "/profile2.jpg"}
		]
	}`

	var credits tmdb.Credits
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&credits)
	require.NoError(t, err)
	require.Len(t, credits.Cast, 1)
	assert.Equal(t, "Actor One", credits.Cast[0].Name)
	assert.Equal(t, "Hero", credits.Cast[0].Character)
	assert.Equal(t, 1, credits.Cast[0].Order)
	require.Len(t, credits.Crew, 1)
	assert.Equal(t, "Director One", credits.Crew[0].Name)
	assert.Equal(t, "Director", credits.Crew[0].Job)
}

func TestKeywords_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"keywords": [
			{"id": 1, "name": "action"},
			{"id": 2, "name": "adventure"}
		]
	}`

	var keywords tmdb.Keywords
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&keywords)
	require.NoError(t, err)
	require.Len(t, keywords.Keywords, 2)
	assert.Equal(t, "action", keywords.Keywords[0].Name)
	assert.Equal(t, "adventure", keywords.Keywords[1].Name)
}

func TestFindResults_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"movie_results": [{"id": 123, "title": "Found Movie"}],
		"tv_results": [{"id": 456, "name": "Found Show"}],
		"person_results": [{"id": 789, "name": "Found Person"}]
	}`

	var results tmdb.FindResults
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&results)
	require.NoError(t, err)
	require.Len(t, results.MovieResults, 1)
	require.Len(t, results.TVResults, 1)
	require.Len(t, results.PersonResults, 1)
	assert.Equal(t, "Found Movie", results.MovieResults[0].Title)
	assert.Equal(t, "Found Show", results.TVResults[0].Name)
	assert.Equal(t, "Found Person", results.PersonResults[0].Name)
}

func TestSeasonDetails_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"_id": "season-123",
		"air_date": "2024-01-15",
		"name": "Season 1",
		"overview": "First season",
		"poster_path": "/poster.jpg",
		"season_number": 1,
		"episodes": [
			{"id": 1, "name": "Episode 1", "episode_number": 1}
		]
	}`

	var season tmdb.SeasonDetails
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&season)
	require.NoError(t, err)
	assert.Equal(t, "season-123", season.ID)
	assert.Equal(t, "2024-01-15", season.AirDate)
	assert.Equal(t, "Season 1", season.Name)
	assert.Equal(t, 1, season.SeasonNumber)
	require.Len(t, season.Episodes, 1)
}

func TestEpisodeDetails_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 500,
		"name": "Episode 5",
		"overview": "A great episode",
		"air_date": "2024-02-15",
		"episode_number": 5,
		"season_number": 2,
		"still_path": "/still.jpg",
		"vote_average": 9.5,
		"vote_count": 500,
		"runtime": 45,
		"production_code": "TST01",
		"show_id": 456,
		"crew": [],
		"guest_stars": []
	}`

	var episode tmdb.EpisodeDetails
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&episode)
	require.NoError(t, err)
	assert.Equal(t, 500, episode.ID)
	assert.Equal(t, "Episode 5", episode.Name)
	assert.Equal(t, 5, episode.EpisodeNumber)
	assert.Equal(t, 2, episode.SeasonNumber)
	assert.Equal(t, 9.5, episode.VoteAverage)
	assert.Equal(t, 500, episode.VoteCount)
	assert.Equal(t, 45, episode.Runtime)
	assert.Equal(t, "TST01", episode.ProductionCode)
	assert.Equal(t, 456, episode.ShowID)
}

func TestExternalIDs_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"imdb_id": "tt1234567",
		"tvdb_id": 12345,
		"facebook_id": "facebook",
		"instagram_id": "instagram",
		"twitter_id": "twitter"
	}`

	var ids tmdb.ExternalIDs
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&ids)
	require.NoError(t, err)
	assert.Equal(t, "tt1234567", ids.IMDbID)
	assert.Equal(t, 12345, ids.TVDBID)
}

func TestCollection_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 100,
		"name": "Collection Name",
		"poster_path": "/collection-poster.jpg",
		"backdrop_path": "/collection-backdrop.jpg"
	}`

	var collection tmdb.Collection
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&collection)
	require.NoError(t, err)
	assert.Equal(t, 100, collection.ID)
	assert.Equal(t, "Collection Name", collection.Name)
	assert.Equal(t, "/collection-poster.jpg", collection.PosterPath)
	assert.Equal(t, "/collection-backdrop.jpg", collection.BackdropPath)
}

func TestNetwork_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 1,
		"name": "NBC",
		"logo_path": "/network-logo.png",
		"origin_country": "US"
	}`

	var network tmdb.Network
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&network)
	require.NoError(t, err)
	assert.Equal(t, 1, network.ID)
	assert.Equal(t, "NBC", network.Name)
	assert.Equal(t, "/network-logo.png", network.LogoPath)
	assert.Equal(t, "US", network.OriginCountry)
}

func TestConfiguration_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"images": {
			"base_url": "http://image.tmdb.org/",
			"secure_base_url": "https://image.tmdb.org/",
			"backdrop_sizes": ["w300", "w780", "w1280"],
			"logo_sizes": ["w45", "w92", "w154"],
			"poster_sizes": ["w92", "w185", "w500", "w780"],
			"profile_sizes": ["w45", "w185"],
			"still_sizes": ["w92", "w185"]
		}
	}`

	var config tmdb.Configuration
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&config)
	require.NoError(t, err)
	assert.Equal(t, "http://image.tmdb.org/", config.Images.BaseURL)
	assert.Equal(t, "https://image.tmdb.org/", config.Images.SecureBaseURL)
	assert.Equal(t, []string{"w300", "w780", "w1280"}, config.Images.BackdropSizes)
	assert.Equal(t, []string{"w92", "w185", "w500", "w780"}, config.Images.PosterSizes)
}

func TestSeasonSummary_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 100,
		"name": "Season 1",
		"overview": "First season",
		"season_number": 1,
		"episode_count": 10,
		"poster_path": "/season1.jpg",
		"air_date": "2024-01-15"
	}`

	var season tmdb.SeasonSummary
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&season)
	require.NoError(t, err)
	assert.Equal(t, 100, season.ID)
	assert.Equal(t, "Season 1", season.Name)
	assert.Equal(t, 1, season.SeasonNumber)
	assert.Equal(t, 10, season.EpisodeCount)
}

func TestGenre_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{"id": 1, "name": "Action"}`

	var genre tmdb.Genre
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&genre)
	require.NoError(t, err)
	assert.Equal(t, 1, genre.ID)
	assert.Equal(t, "Action", genre.Name)
}

func TestCastMember_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 1,
		"name": "Actor Name",
		"character": "Hero",
		"profile_path": "/profile.jpg",
		"order": 1,
		"cast_id": 100,
		"credit_id": "credit-123"
	}`

	var cast tmdb.CastMember
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&cast)
	require.NoError(t, err)
	assert.Equal(t, 1, cast.ID)
	assert.Equal(t, "Actor Name", cast.Name)
	assert.Equal(t, "Hero", cast.Character)
	assert.Equal(t, 1, cast.Order)
}

func TestCrewMember_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 2,
		"name": "Director Name",
		"job": "Director",
		"department": "Directing",
		"profile_path": "/crew.jpg",
		"credit_id": "credit-456"
	}`

	var crew tmdb.CrewMember
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&crew)
	require.NoError(t, err)
	assert.Equal(t, 2, crew.ID)
	assert.Equal(t, "Director Name", crew.Name)
	assert.Equal(t, "Director", crew.Job)
	assert.Equal(t, "Directing", crew.Department)
}

func TestPersonSummary_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 789,
		"name": "Person Name",
		"profile_path": "/person.jpg"
	}`

	var person tmdb.PersonSummary
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&person)
	require.NoError(t, err)
	assert.Equal(t, 789, person.ID)
	assert.Equal(t, "Person Name", person.Name)
}

func TestCompany_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 1,
		"name": "Studio Name",
		"logo_path": "/logo.png",
		"origin_country": "US"
	}`

	var company tmdb.Company
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&company)
	require.NoError(t, err)
	assert.Equal(t, 1, company.ID)
	assert.Equal(t, "Studio Name", company.Name)
	assert.Equal(t, "US", company.OriginCountry)
}

func TestCountry_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"iso_3166_1": "US",
		"name": "United States"
	}`

	var country tmdb.Country
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&country)
	require.NoError(t, err)
	assert.Equal(t, "US", country.ISO3166_1)
	assert.Equal(t, "United States", country.Name)
}

func TestLanguage_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"iso_639_1": "en",
		"name": "English",
		"english_name": "English"
	}`

	var lang tmdb.Language
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&lang)
	require.NoError(t, err)
	assert.Equal(t, "en", lang.ISO639_1)
	assert.Equal(t, "English", lang.Name)
}

func TestReleaseDates_Parsing(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 123,
		"results": [
			{
				"iso_3166_1": "US",
				"release_dates": [
					{"certification": "PG-13", "release_date": "2024-01-15T00:00:00Z", "type": 3}
				]
			}
		]
	}`

	var rd tmdb.ReleaseDates
	err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&rd)
	require.NoError(t, err)
	assert.Equal(t, 123, rd.ID)
	require.Len(t, rd.Results, 1)
	assert.Equal(t, "US", rd.Results[0].ISO3166_1)
	require.Len(t, rd.Results[0].ReleaseDates, 1)
	assert.Equal(t, "PG-13", rd.Results[0].ReleaseDates[0].Certification)
	assert.Equal(t, 3, rd.Results[0].ReleaseDates[0].Type)
}

// Test Client creation
func TestNewClient(t *testing.T) {
	t.Parallel()
	c := tmdb.New("test-api-key")
	assert.NotNil(t, c)
}

func TestNewWithHTTPClient(t *testing.T) {
	t.Parallel()
	c := tmdb.NewWithHTTPClient("test-api-key", nil)
	assert.NotNil(t, c)
}

// Test GetImageURL with and without config
func TestGetImageURL_NoConfig(t *testing.T) {
	t.Parallel()
	c := tmdb.New("test-key")
	url := c.GetImageURL("/poster.jpg", "w500")
	assert.Equal(t, "", url)
}

func TestGetImageURL_EmptyPath(t *testing.T) {
	t.Parallel()
	c := tmdb.New("test-key")
	// Even with nil config behavior, empty path should return empty
	url := c.GetImageURL("", "w500")
	assert.Equal(t, "", url)
}
