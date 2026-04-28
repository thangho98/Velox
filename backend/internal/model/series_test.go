package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeries_JSONMarshaling(t *testing.T) {
	t.Parallel()

	series := Series{
		ID:             1,
		LibraryID:      10,
		Title:          "Test Series",
		SortTitle:      "Test Series",
		TmdbID:         ptrInt64(12345),
		AnilistID:      ptrInt64(98765),
		RomajiTitle:    "テストシリーズ",
		Studio:         "Test Studio",
		Overview:       "A test series overview",
		Status:         "Returning Series",
		Network:        "Netflix",
		FirstAirDate:   "2024-01-15",
		MetadataLocked: false,
		CreatedAt:      "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(series)
	require.NoError(t, err)

	var got Series
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, series.ID, got.ID)
	assert.Equal(t, series.LibraryID, got.LibraryID)
	assert.Equal(t, series.Title, got.Title)
	assert.Equal(t, series.SortTitle, got.SortTitle)
	assert.Equal(t, series.TmdbID, got.TmdbID)
	assert.Equal(t, series.AnilistID, got.AnilistID)
	assert.Equal(t, series.RomajiTitle, got.RomajiTitle)
	assert.Equal(t, series.Studio, got.Studio)
	assert.Equal(t, series.Overview, got.Overview)
	assert.Equal(t, series.Status, got.Status)
	assert.Equal(t, series.Network, got.Network)
	assert.Equal(t, series.FirstAirDate, got.FirstAirDate)
	assert.Equal(t, series.MetadataLocked, got.MetadataLocked)
}

func TestSeries_JSONOmitsPathFields(t *testing.T) {
	t.Parallel()

	series := Series{
		ID:        1,
		Title:     "Test",
		CreatedAt: "2024-01-01",
		UpdatedAt: "2024-01-01",
	}

	data, err := json.Marshal(series)
	require.NoError(t, err)

	var got map[string]interface{}
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	// Path fields have json:"-" so should not appear
	_, hasPosterPath := got["posterPath"]
	_, hasBackdropPath := got["backdropPath"]
	_, hasLogoPath := got["logoPath"]
	_, hasThumbPath := got["thumbPath"]

	assert.False(t, hasPosterPath, "posterPath should not be in JSON")
	assert.False(t, hasBackdropPath, "backdropPath should not be in JSON")
	assert.False(t, hasLogoPath, "logoPath should not be in JSON")
	assert.False(t, hasThumbPath, "thumbPath should not be in JSON")
}

func TestSeries_UnmarshalWithPointers(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"id": 5,
		"library_id": 2,
		"title": "My Series",
		"sort_title": "My Series",
		"tmdb_id": 12345,
		"imdb_id": "tt1234567",
		"tvdb_id": 98765,
		"anilist_id": 55555,
		"romaji_title": "マイシリーズ",
		"studio": "Studio A",
		"overview": "Overview text",
		"status": "Ended",
		"network": "HBO",
		"first_air_date": "2023-06-15",
		"metadata_locked": true,
		"created_at": "2023-01-01T00:00:00Z",
		"updated_at": "2023-06-15T00:00:00Z"
	}`)

	var series Series
	err := json.Unmarshal(data, &series)
	require.NoError(t, err)

	assert.Equal(t, int64(5), series.ID)
	assert.Equal(t, int64(2), series.LibraryID)
	require.NotNil(t, series.TmdbID)
	assert.Equal(t, int64(12345), *series.TmdbID)
	require.NotNil(t, series.ImdbID)
	assert.Equal(t, "tt1234567", *series.ImdbID)
	require.NotNil(t, series.TvdbID)
	assert.Equal(t, int64(98765), *series.TvdbID)
	require.NotNil(t, series.AnilistID)
	assert.Equal(t, int64(55555), *series.AnilistID)
	assert.Equal(t, "Ended", series.Status)
	assert.True(t, series.MetadataLocked)
}

func TestSeason_JSONMarshaling(t *testing.T) {
	t.Parallel()

	season := Season{
		ID:           1,
		SeriesID:     10,
		SeasonNumber: 1,
		Title:        "Season 1",
		Overview:     "First season",
		EpisodeCount: 12,
		CreatedAt:    "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(season)
	require.NoError(t, err)

	var got Season
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, season.ID, got.ID)
	assert.Equal(t, season.SeriesID, got.SeriesID)
	assert.Equal(t, season.SeasonNumber, got.SeasonNumber)
	assert.Equal(t, season.Title, got.Title)
	assert.Equal(t, season.Overview, got.Overview)
	assert.Equal(t, season.EpisodeCount, got.EpisodeCount)
}

func TestSeason_JSONOmitsPosterPath(t *testing.T) {
	t.Parallel()

	season := Season{
		ID:           1,
		SeasonNumber: 1,
		CreatedAt:    "2024-01-01",
	}

	data, err := json.Marshal(season)
	require.NoError(t, err)

	var got map[string]interface{}
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	_, exists := got["posterPath"]
	assert.False(t, exists, "posterPath should not be in JSON")
}

func TestEpisode_JSONMarshaling(t *testing.T) {
	t.Parallel()

	ep := Episode{
		ID:            1,
		SeriesID:      10,
		SeasonID:      5,
		MediaID:       100,
		EpisodeNumber: 3,
		Title:         "Episode 3",
		Overview:      "Third episode",
		AirDate:       "2024-01-15",
		Duration:      3600.5,
		CreatedAt:     "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(ep)
	require.NoError(t, err)

	var got Episode
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, ep.ID, got.ID)
	assert.Equal(t, ep.SeriesID, got.SeriesID)
	assert.Equal(t, ep.SeasonID, got.SeasonID)
	assert.Equal(t, ep.MediaID, got.MediaID)
	assert.Equal(t, ep.EpisodeNumber, got.EpisodeNumber)
	assert.Equal(t, ep.Title, got.Title)
	assert.Equal(t, ep.Overview, got.Overview)
	assert.Equal(t, ep.AirDate, got.AirDate)
	assert.Equal(t, ep.Duration, got.Duration)
}

func TestEpisode_OmitsStillPath(t *testing.T) {
	t.Parallel()

	ep := Episode{
		ID:            1,
		EpisodeNumber: 1,
		Title:         "Test",
		CreatedAt:     "2024-01-01",
	}

	data, err := json.Marshal(ep)
	require.NoError(t, err)

	var got map[string]interface{}
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	_, exists := got["stillPath"]
	assert.False(t, exists, "stillPath should not be in JSON")
}

func TestEpisodeWithMedia_JSONMarshaling(t *testing.T) {
	t.Parallel()

	ewm := EpisodeWithMedia{
		Episode: Episode{
			ID:            1,
			SeriesID:      10,
			SeasonID:      5,
			MediaID:       100,
			EpisodeNumber: 3,
			Title:         "Episode 3",
		},
		Media: Media{
			ID:        100,
			Title:     "Episode 3 Media",
			MediaType: "episode",
		},
		PrimaryFile: &MediaFile{
			ID:       200,
			MediaID:  100,
			FilePath: "/path/to/file.mp4",
			Duration: 3600,
		},
	}

	data, err := json.Marshal(ewm)
	require.NoError(t, err)

	var got EpisodeWithMedia
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, ewm.Episode.ID, got.Episode.ID)
	assert.Equal(t, ewm.Episode.EpisodeNumber, got.Episode.EpisodeNumber)
	assert.Equal(t, ewm.Media.ID, got.Media.ID)
	require.NotNil(t, got.PrimaryFile)
	assert.Equal(t, ewm.PrimaryFile.ID, got.PrimaryFile.ID)
	assert.Equal(t, ewm.PrimaryFile.FilePath, got.PrimaryFile.FilePath)
}

func TestSeriesWithSeasons_JSONMarshaling(t *testing.T) {
	t.Parallel()

	sws := SeriesWithSeasons{
		Series: Series{
			ID:    1,
			Title: "Test Series",
		},
		Seasons: []Season{
			{ID: 10, SeriesID: 1, SeasonNumber: 1, Title: "Season 1"},
			{ID: 20, SeriesID: 1, SeasonNumber: 2, Title: "Season 2"},
		},
	}

	data, err := json.Marshal(sws)
	require.NoError(t, err)

	var got SeriesWithSeasons
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, sws.Series.ID, got.Series.ID)
	assert.Equal(t, sws.Series.Title, got.Series.Title)
	require.Len(t, got.Seasons, 2)
	assert.Equal(t, sws.Seasons[0].SeasonNumber, got.Seasons[0].SeasonNumber)
	assert.Equal(t, sws.Seasons[1].SeasonNumber, got.Seasons[1].SeasonNumber)
}

func TestSeriesListItem_JSONMarshaling(t *testing.T) {
	t.Parallel()

	item := SeriesListItem{
		ID:           1,
		LibraryID:    10,
		Title:        "Test Series",
		Status:       "Returning Series",
		Genres:       []string{"Action", "Drama"},
		SeasonCount:  3,
		EpisodeCount: 36,
		CreatedAt:    "2024-01-01T00:00:00Z",
		UpdatedAt:    "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

	var got SeriesListItem
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, item.ID, got.ID)
	assert.Equal(t, item.Title, got.Title)
	assert.Equal(t, item.Status, got.Status)
	assert.Equal(t, item.Genres, got.Genres)
	assert.Equal(t, item.SeasonCount, got.SeasonCount)
	assert.Equal(t, item.EpisodeCount, got.EpisodeCount)
}

func TestSeriesMetadataEditRequest_PartialUpdate(t *testing.T) {
	t.Parallel()

	// Test with only title being updated
	title := "New Title"
	data := []byte(`{"title": "New Title"}`)

	var req SeriesMetadataEditRequest
	err := json.Unmarshal(data, &req)
	require.NoError(t, err)

	require.NotNil(t, req.Title)
	assert.Equal(t, title, *req.Title)
	assert.Nil(t, req.Overview)
	assert.Nil(t, req.Status)
	assert.Nil(t, req.MetadataLocked)
}

func TestEpisodeMetadataEditRequest_PartialUpdate(t *testing.T) {
	t.Parallel()

	episodeNum := 5
	data := []byte(`{"episode_number": 5}`)

	var req EpisodeMetadataEditRequest
	err := json.Unmarshal(data, &req)
	require.NoError(t, err)

	require.NotNil(t, req.EpisodeNumber)
	assert.Equal(t, episodeNum, *req.EpisodeNumber)
	assert.Nil(t, req.Title)
	assert.Nil(t, req.Overview)
}

func TestSeriesListFilter_JSONMarshaling(t *testing.T) {
	t.Parallel()

	filter := SeriesListFilter{
		LibraryID: 1,
		Search:    "test",
		Genre:     "Action",
		Year:      "2024",
		Sort:      "newest",
		Limit:     20,
		Offset:    0,
		StartChar: "A",
	}

	data, err := json.Marshal(filter)
	require.NoError(t, err)

	var got SeriesListFilter
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, filter.LibraryID, got.LibraryID)
	assert.Equal(t, filter.Search, got.Search)
	assert.Equal(t, filter.Genre, got.Genre)
	assert.Equal(t, filter.Year, got.Year)
	assert.Equal(t, filter.Sort, got.Sort)
	assert.Equal(t, filter.Limit, got.Limit)
	assert.Equal(t, filter.Offset, got.Offset)
	assert.Equal(t, filter.StartChar, got.StartChar)
}

// Helper function for tests
func ptrInt64(v int64) *int64 {
	return &v
}
