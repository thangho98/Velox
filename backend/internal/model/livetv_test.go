package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLivePlaylist_JSONMarshaling(t *testing.T) {
	t.Parallel()

	now := time.Now()
	playlist := LivePlaylist{
		ID:                1,
		Name:              "Test Playlist",
		URL:               "https://example.com/playlist.m3u",
		EpgURL:            "https://example.com/epg.xml",
		LastSyncedAt:      &now,
		SyncIntervalHours: 6,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	data, err := json.Marshal(playlist)
	require.NoError(t, err)

	var got LivePlaylist
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, playlist.ID, got.ID)
	assert.Equal(t, playlist.Name, got.Name)
	assert.Equal(t, playlist.URL, got.URL)
	assert.Equal(t, playlist.EpgURL, got.EpgURL)
	assert.Equal(t, playlist.SyncIntervalHours, got.SyncIntervalHours)
	assert.Equal(t, playlist.IsActive, got.IsActive)
}

func TestLivePlaylist_JSONOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	playlist := LivePlaylist{
		ID:                1,
		Name:              "Test Playlist",
		URL:               "https://example.com/playlist.m3u",
		SyncIntervalHours: 6,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		// LastSyncedAt is nil - should be omitted
	}

	data, err := json.Marshal(playlist)
	require.NoError(t, err)

	var got map[string]interface{}
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	_, exists := got["lastSyncedAt"]
	assert.False(t, exists, "lastSyncedAt should be omitted when nil")
}

func TestLiveChannel_JSONMarshaling(t *testing.T) {
	t.Parallel()

	ch := LiveChannel{
		ID:         1,
		PlaylistID: 10,
		ChannelID:  "ch123",
		Name:       "Test Channel",
		Logo:       "https://example.com/logo.png",
		GroupTitle: "Sports",
		StreamURL:  "https://example.com/stream.m3u8",
		Country:    "US",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(ch)
	require.NoError(t, err)

	var got LiveChannel
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, ch.ID, got.ID)
	assert.Equal(t, ch.PlaylistID, got.PlaylistID)
	assert.Equal(t, ch.ChannelID, got.ChannelID)
	assert.Equal(t, ch.Name, got.Name)
	assert.Equal(t, ch.Logo, got.Logo)
	assert.Equal(t, ch.GroupTitle, got.GroupTitle)
	assert.Equal(t, ch.StreamURL, got.StreamURL)
	assert.Equal(t, ch.Country, got.Country)
	assert.Equal(t, ch.IsActive, got.IsActive)
}

func TestLiveChannel_JSONTags(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"id": 5,
		"playlistId": 2,
		"channelId": "espn",
		"name": "ESPN",
		"logo": "https://example.com/espn.png",
		"groupTitle": "Sports",
		"streamUrl": "https://example.com/espn.m3u8",
		"country": "US",
		"isActive": true,
		"createdAt": "2024-01-01T00:00:00Z",
		"updatedAt": "2024-01-01T00:00:00Z"
	}`)

	var ch LiveChannel
	err := json.Unmarshal(data, &ch)
	require.NoError(t, err)

	assert.Equal(t, int64(5), ch.ID)
	assert.Equal(t, int64(2), ch.PlaylistID)
	assert.Equal(t, "espn", ch.ChannelID)
	assert.Equal(t, "ESPN", ch.Name)
	assert.Equal(t, "https://example.com/espn.png", ch.Logo)
}

func TestLiveProgram_JSONMarshaling(t *testing.T) {
	t.Parallel()

	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	prog := LiveProgram{
		ID:          1,
		ChannelID:   5,
		Title:       "Morning Show",
		Description: "Today's morning show",
		StartTime:   start,
		EndTime:     end,
	}

	data, err := json.Marshal(prog)
	require.NoError(t, err)

	var got LiveProgram
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, prog.ID, got.ID)
	assert.Equal(t, prog.ChannelID, got.ChannelID)
	assert.Equal(t, prog.Title, got.Title)
	assert.Equal(t, prog.Description, got.Description)
}

func TestLiveProgram_UnmarshalWithSnakeCase(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"id": 10,
		"channel_id": 3,
		"title": "News",
		"description": "Evening news",
		"start_time": "2024-01-15T18:00:00Z",
		"end_time": "2024-01-15T19:00:00Z"
	}`)

	var prog LiveProgram
	err := json.Unmarshal(data, &prog)
	require.NoError(t, err)

	assert.Equal(t, int64(10), prog.ID)
	assert.Equal(t, int64(3), prog.ChannelID)
	assert.Equal(t, "News", prog.Title)
	assert.Equal(t, "Evening news", prog.Description)
}

func TestLiveProgram_OmitsOptionalFields(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"id": 1,
		"channel_id": 5,
		"title": "Test Show",
		"start_time": "2024-01-15T10:00:00Z",
		"end_time": "2024-01-15T11:00:00Z"
	}`)

	var prog LiveProgram
	err := json.Unmarshal(data, &prog)
	require.NoError(t, err)

	assert.Equal(t, int64(1), prog.ID)
	assert.Equal(t, "", prog.Description, "optional description should be empty string")
}
