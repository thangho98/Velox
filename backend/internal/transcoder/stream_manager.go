package transcoder

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/thawng/velox/internal/hls"
	"github.com/thawng/velox/internal/logger"
	"github.com/thawng/velox/internal/model"
)

// Manager manages per-viewer HLS V2 streaming sessions.
// Each session owns a single long-running jellyfin-ffmpeg process that writes
// fMP4 segments on demand.  Sessions are garbage-collected after idleTimeout.
type Manager struct {
	mu          sync.Mutex
	sessions    map[hls.SessionKey]*Session
	gcInterval  time.Duration
	idleTimeout time.Duration
	baseOutDir  string
	hwAccel     string
	log         *slog.Logger
}

// NewManager creates a Manager and starts background session GC.
func NewManager(baseOutDir string, hwAccel string) *Manager {
	m := &Manager{
		sessions:    make(map[hls.SessionKey]*Session),
		gcInterval:  1 * time.Minute,
		idleTimeout: 10 * time.Minute,
		baseOutDir:  baseOutDir,
		hwAccel:     hwAccel,
		log:         logger.NewWith("stream-mgr"),
	}
	go m.gcLoop()
	return m
}

// GetOrCreate returns an existing session for the given key, or creates a new one.
// Enforces 1 session per StreamSessionID: if the same viewer already has a session
// with different parameters (e.g. quality change) the old session is closed first.
func (m *Manager) GetOrCreate(ctx context.Context, key hls.SessionKey, inputPath string, totalDur float64, audioTracks []model.AudioTrack) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Enforce 1 session ID per viewer.
	for k, v := range m.sessions {
		if k.StreamSessionID == key.StreamSessionID && k != key {
			v.Close()
			delete(m.sessions, k)
			m.log.Info("Closed older session for StreamSessionID", "id", k.StreamSessionID)
		}
	}

	if sess, ok := m.sessions[key]; ok {
		sess.Touch()
		return sess, nil
	}

	outDir := filepath.Join(m.baseOutDir, fmt.Sprintf("%d", key.MediaID))
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, err
	}

	sess := &Session{
		Key:         key,
		OutputDir:   outDir,
		InputPath:   inputPath,
		HwAccel:     m.hwAccel,
		TotalDur:    totalDur,
		AudioTracks: audioTracks,
		SegLength:   6.0,
		log:         logger.NewWith("session_v2", "ss", key.StreamSessionID),
		extinfMap:   make(map[string]map[int]float64),
	}
	sess.Touch()
	m.sessions[key] = sess

	return sess, nil
}

// CancelByMediaID closes and removes all sessions for a media item.
func (m *Manager) CancelByMediaID(mediaID int64) int {
	if mediaID <= 0 {
		return 0
	}

	return m.cancelWhere(func(key hls.SessionKey) bool {
		return key.MediaID == mediaID
	})
}

// CancelByStreamSessionID closes and removes all sessions for a viewer session.
func (m *Manager) CancelByStreamSessionID(streamSessionID string) int {
	if streamSessionID == "" {
		return 0
	}

	return m.cancelWhere(func(key hls.SessionKey) bool {
		return key.StreamSessionID == streamSessionID
	})
}

func (m *Manager) cancelWhere(matches func(hls.SessionKey) bool) int {
	m.mu.Lock()
	toClose := make([]*Session, 0)
	for key, sess := range m.sessions {
		if matches(key) {
			toClose = append(toClose, sess)
			delete(m.sessions, key)
		}
	}
	m.mu.Unlock()

	for _, sess := range toClose {
		sess.Close()
	}

	return len(toClose)
}

func (m *Manager) gcLoop() {
	ticker := time.NewTicker(m.gcInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, sess := range m.sessions {
			sess.mu.Lock()
			idleDur := now.Sub(sess.lastAccess)
			sess.mu.Unlock()

			if idleDur > m.idleTimeout {
				m.log.Info("GC closed idle session", "key", key)
				sess.Close()
				delete(m.sessions, key)
			}
		}
		m.mu.Unlock()
	}
}
