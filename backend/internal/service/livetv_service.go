package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service/m3uparser"
)

type LiveTVService struct {
	repo   *repository.LiveTVRepo
	parser *m3uparser.Parser
}

func NewLiveTVService(repo *repository.LiveTVRepo, parser *m3uparser.Parser) *LiveTVService {
	return &LiveTVService{
		repo:   repo,
		parser: parser,
	}
}

// RegisterTasks registers the Live TV background sync tasks with the scheduler
func (s *LiveTVService) RegisterTasks(scheduler *Scheduler) {
	scheduler.Register("livetv-sync", 24*time.Hour, func(ctx context.Context) error {
		return s.SyncAllPlaylists(ctx)
	})
}

// SyncAllPlaylists retrieves all active playlists and syncs their channels
func (s *LiveTVService) SyncAllPlaylists(ctx context.Context) error {
	playlists, err := s.repo.GetActivePlaylists(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active playlists: %w", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(playlists))

	for _, p := range playlists {
		wg.Add(1)
		go func(playlist *model.LivePlaylist) {
			defer wg.Done()

			// Check sync interval if needed (omitted for strict forced sync right now)
			// Pass the parent context to properly support cancellation
			err := s.SyncPlaylist(ctx, playlist.ID)
			if err != nil {
				log.Printf("[LiveTV] Failed to sync playlist %d (%s): %v", playlist.ID, playlist.Name, err)
				errs <- fmt.Errorf("playlist %d: %w", playlist.ID, err)
			} else {
				log.Printf("[LiveTV] Successfully synced playlist %d (%s)", playlist.ID, playlist.Name)
			}
		}(p)
	}

	wg.Wait()
	close(errs)

	var errMsgs []string
	for e := range errs {
		errMsgs = append(errMsgs, e.Error())
	}
	if len(errMsgs) > 0 {
		return fmt.Errorf("sync completed with errors: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

// SyncPlaylist fetches channels for a specific playlist and updates the database
func (s *LiveTVService) SyncPlaylist(ctx context.Context, playlistID int64) error {
	// 1. Fetch playlist from db
	playlist, err := s.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("playlist with ID %d not found or fetch error: %w", playlistID, err)
	}

	// 2. Fetch the actual M3U channels
	var channels []*model.LiveChannel
	var epgURL string
	if strings.HasPrefix(playlist.URL, "http://") || strings.HasPrefix(playlist.URL, "https://") {
		// Needs context timeout, 5 min should be plenty for even huge playlists
		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		channels, epgURL, err = s.parser.ParseURL(reqCtx, playlist.URL)
	} else {
		// Treat as local file path
		channels, epgURL, err = s.parser.ParseFile(playlist.URL)
	}

	if err != nil {
		return fmt.Errorf("parse playlist failed: %w", err)
	}

	if len(channels) == 0 {
		return fmt.Errorf("no channels found in playlist")
	}

	// 2.5 Update playlist with EPG URL if found and not manually overridden
	if epgURL != "" && playlist.EpgURL == "" {
		playlist.EpgURL = epgURL
		if updateErr := s.repo.UpdatePlaylist(ctx, playlist); updateErr != nil {
			log.Printf("[LiveTV] Failed to update parsed EpgURL for playlist %d: %v", playlistID, updateErr)
		}
	}

	// 3 & 4. Deactivate and Bulk Upsert the new channels in a single transaction
	if err := s.repo.SyncChannelsTransaction(ctx, playlistID, channels); err != nil {
		return fmt.Errorf("transaction sync channels failed: %w", err)
	}

	// 5. Update sync timestamp
	if err := s.repo.UpdatePlaylistSyncedAt(ctx, playlistID); err != nil {
		log.Printf("[LiveTV] Failed to update synced_at for playlist %d: %v", playlistID, err)
	}

	return nil
}

// ToggleChannelHidden toggles a channel's hidden status
func (s *LiveTVService) ToggleChannelHidden(ctx context.Context, id int64) (bool, error) {
	return s.repo.ToggleChannelHidden(ctx, id)
}
