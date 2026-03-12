package playback

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// TranscodeSession tracks an active FFmpeg process
type TranscodeSession struct {
	ID        string         `json:"id"`
	MediaID   int            `json:"media_id"`
	UserID    int            `json:"user_id"`
	Process   *os.Process    `json:"-"`
	PID       int            `json:"pid"`
	StartedAt time.Time      `json:"started_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Progress  float64        `json:"progress"` // 0-100
	Method    PlaybackMethod `json:"method"`
	Status    string         `json:"status"` // running, completed, failed, killed
	OutputDir string         `json:"output_dir"`
}

// SessionManager manages active transcode sessions
type SessionManager struct {
	sessions    map[string]*TranscodeSession
	mu          sync.RWMutex
	maxSessions int
}

// NewSessionManager creates a new session manager
func NewSessionManager(maxSessions int) *SessionManager {
	if maxSessions <= 0 {
		maxSessions = 2 // Default limit
	}
	return &SessionManager{
		sessions:    make(map[string]*TranscodeSession),
		maxSessions: maxSessions,
	}
}

// CreateSession creates a new transcode session
func (m *SessionManager) CreateSession(mediaID, userID int, method PlaybackMethod, outputDir string) (*TranscodeSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check session limit
	if len(m.sessions) >= m.maxSessions {
		// Clean up old sessions first
		m.cleanupOldSessions()

		if len(m.sessions) >= m.maxSessions {
			return nil, fmt.Errorf("max concurrent transcode sessions reached (%d)", m.maxSessions)
		}
	}

	// Generate session ID
	var randBytes [8]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		return nil, fmt.Errorf("generating session id: %w", err)
	}
	id := fmt.Sprintf("%d-%d-%s", userID, mediaID, hex.EncodeToString(randBytes[:]))

	session := &TranscodeSession{
		ID:        id,
		MediaID:   mediaID,
		UserID:    userID,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
		Method:    method,
		Status:    "starting",
		OutputDir: outputDir,
	}

	m.sessions[id] = session
	return session, nil
}

// AttachProcess attaches an FFmpeg process to a session
func (m *SessionManager) AttachProcess(sessionID string, cmd *exec.Cmd) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if cmd.Process != nil {
		session.Process = cmd.Process
		session.PID = cmd.Process.Pid
		session.Status = "running"
	}

	return nil
}

// UpdateProgress updates session progress
func (m *SessionManager) UpdateProgress(sessionID string, progress float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.Progress = progress
		session.UpdatedAt = time.Now()
	}
}

// CompleteSession marks a session as completed
func (m *SessionManager) CompleteSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.Status = "completed"
		session.Progress = 100
		session.UpdatedAt = time.Now()
	}
}

// FailSession marks a session as failed
func (m *SessionManager) FailSession(sessionID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.Status = "failed"
		session.UpdatedAt = time.Now()
	}
}

// KillSession kills a transcode session
func (m *SessionManager) KillSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Process != nil {
		// Try graceful termination first
		err := session.Process.Signal(os.Interrupt)
		if err != nil {
			// Force kill
			err = session.Process.Kill()
		}

		session.Status = "killed"
		session.UpdatedAt = time.Now()
		return err
	}

	return nil
}

// RemoveSession removes a session from tracking
func (m *SessionManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
}

// GetSession gets a session by ID
func (m *SessionManager) GetSession(sessionID string) (*TranscodeSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	return session, exists
}

// GetUserSessions gets all sessions for a user
func (m *SessionManager) GetUserSessions(userID int) []*TranscodeSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*TranscodeSession
	for _, session := range m.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}
	return result
}

// GetMediaSession gets active session for media+user combination
func (m *SessionManager) GetMediaSession(mediaID, userID int) (*TranscodeSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.MediaID == mediaID && session.UserID == userID && session.Status == "running" {
			return session, true
		}
	}
	return nil, false
}

// ListActiveSessions returns all active sessions
func (m *SessionManager) ListActiveSessions() []*TranscodeSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*TranscodeSession
	for _, session := range m.sessions {
		if session.Status == "running" || session.Status == "starting" {
			result = append(result, session)
		}
	}
	return result
}

// cleanupOldSessions removes stale sessions
func (m *SessionManager) cleanupOldSessions() {
	now := time.Now()
	for id, session := range m.sessions {
		// Remove completed/failed sessions older than 1 hour
		if session.Status == "completed" || session.Status == "failed" || session.Status == "killed" {
			if now.Sub(session.UpdatedAt) > time.Hour {
				delete(m.sessions, id)
			}
		}
		// Kill stale running sessions (no update for 5 minutes)
		if session.Status == "running" && now.Sub(session.UpdatedAt) > 5*time.Minute {
			if session.Process != nil {
				session.Process.Kill()
			}
			session.Status = "killed"
			delete(m.sessions, id)
		}
	}
}

// CleanupOnStartup kills any orphaned FFmpeg processes
func (m *SessionManager) CleanupOnStartup() error {
	// In a real implementation, this would:
	// 1. Check for running FFmpeg processes from previous instance
	// 2. Kill them to prevent resource leaks
	// 3. Clean up temp directories

	// For now, just clear the session map
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make(map[string]*TranscodeSession)
	return nil
}

// StartCleanupTask starts a background task to cleanup stale sessions
func (m *SessionManager) StartCleanupTask(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			m.cleanupOldSessions()
			m.mu.Unlock()
		}
	}
}

// SessionStats holds session statistics
type SessionStats struct {
	Total      int `json:"total"`
	Running    int `json:"running"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Killed     int `json:"killed"`
	MaxAllowed int `json:"max_allowed"`
}

// GetStats returns session statistics
func (m *SessionManager) GetStats() SessionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := SessionStats{MaxAllowed: m.maxSessions}
	for _, session := range m.sessions {
		stats.Total++
		switch session.Status {
		case "running":
			stats.Running++
		case "completed":
			stats.Completed++
		case "failed":
			stats.Failed++
		case "killed":
			stats.Killed++
		}
	}
	return stats
}
