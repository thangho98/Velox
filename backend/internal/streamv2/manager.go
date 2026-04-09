package streamv2

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

type Manager struct {
	mu          sync.Mutex
	sessions    map[hls.SessionKey]*Session
	gcInterval  time.Duration
	idleTimeout time.Duration
	baseOutDir  string
	hwAccel     string
	log         *slog.Logger
}

func NewManager(baseOutDir string, hwAccel string) *Manager {
	m := &Manager{
		sessions:    make(map[hls.SessionKey]*Session),
		gcInterval:  1 * time.Minute,
		idleTimeout: 10 * time.Minute,
		baseOutDir:  baseOutDir,
		hwAccel:     hwAccel,
		log:         logger.NewWith("streamv2"),
	}
	go m.gcLoop()
	return m
}

func (m *Manager) GetOrCreate(ctx context.Context, key hls.SessionKey, inputPath string, totalDur float64, audioTracks []model.AudioTrack) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Enforce 1 session ID per viewer
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

	// Match dir structure of legacy HLS: baseOutDir/hls/<mediaID> (just passing mediaID as part of outDir contextually)
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
