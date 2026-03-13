package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

const (
	activityBufferSize = 256
	activityFlushSize  = 50
	activityFlushEvery = 5 * time.Second
)

// ActivityService provides async activity logging with buffered writes.
type ActivityService struct {
	repo *repository.ActivityRepo
	ch   chan repository.ActivityEntry
	done chan struct{}
	wg   sync.WaitGroup
}

func NewActivityService(repo *repository.ActivityRepo) *ActivityService {
	s := &ActivityService{
		repo: repo,
		ch:   make(chan repository.ActivityEntry, activityBufferSize),
		done: make(chan struct{}),
	}
	s.wg.Add(1)
	go s.flushLoop()
	return s
}

// Log enqueues an activity entry for async insertion. Non-blocking; drops if buffer full.
func (s *ActivityService) Log(userID *int64, action, ip string, mediaID *int64, details string) {
	entry := repository.ActivityEntry{
		UserID:  userID,
		Action:  action,
		MediaID: mediaID,
		Details: details,
		IP:      ip,
	}
	select {
	case s.ch <- entry:
	default:
		// Buffer full, drop entry to avoid blocking the caller
		log.Printf("activity: buffer full, dropping %s event", action)
	}
}

// ListActivity retrieves activity logs with filters.
func (s *ActivityService) ListActivity(ctx context.Context, filter model.ActivityFilter) ([]model.ActivityLog, error) {
	return s.repo.List(ctx, filter)
}

// GetPlaybackStats returns aggregated playback statistics.
func (s *ActivityService) GetPlaybackStats(ctx context.Context) (*model.PlaybackStatsResult, error) {
	return s.repo.PlaybackStats(ctx)
}

// Close flushes remaining entries and stops the background goroutine.
func (s *ActivityService) Close() {
	close(s.done)
	s.wg.Wait()
}

func (s *ActivityService) flushLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(activityFlushEvery)
	defer ticker.Stop()

	var buf []repository.ActivityEntry

	for {
		select {
		case entry := <-s.ch:
			buf = append(buf, entry)
			if len(buf) >= activityFlushSize {
				s.flush(buf)
				buf = nil
			}

		case <-ticker.C:
			if len(buf) > 0 {
				s.flush(buf)
				buf = nil
			}

		case <-s.done:
			// Drain remaining entries from channel
			close(s.ch)
			for entry := range s.ch {
				buf = append(buf, entry)
			}
			if len(buf) > 0 {
				s.flush(buf)
			}
			return
		}
	}
}

func (s *ActivityService) flush(entries []repository.ActivityEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.repo.InsertBatch(ctx, entries); err != nil {
		log.Printf("activity: flush error (%d entries): %v", len(entries), err)
	}
}
