package transcoder

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/thawng/velox/internal/hls"
	"github.com/thawng/velox/internal/model"
)

// Session manages a single viewer's HLS V2 streaming session.
// It owns one long-running jellyfin-ffmpeg process that writes fMP4 segments
// for the requested timeline position. PrimeFromSegment restarts ffmpeg when
// the player seeks outside the current buffer window.
type Session struct {
	Key         hls.SessionKey
	OutputDir   string
	InputPath   string
	HwAccel     string
	TotalDur    float64
	AudioTracks []model.AudioTrack
	SegLength   float64

	mu           sync.Mutex
	cmd          *exec.Cmd
	cmdCancel    context.CancelFunc
	cmdExited    chan struct{}
	startSegment int
	lastAccess   time.Time
	log          *slog.Logger

	extinfMu      sync.RWMutex
	extinfMap     map[string]map[int]float64
	lastRequested int

	restartMu sync.Mutex
}

// Prefix returns the filename prefix used for all HLS output files in this session.
func (s *Session) Prefix() string {
	return hls.BuildPrefix(s.Key)
}

// Touch updates the session's last-access timestamp to prevent GC eviction.
func (s *Session) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastAccess = time.Now()
}

// Close terminates the ffmpeg process and waits for it to exit.
func (s *Session) Close() {
	s.restartMu.Lock()
	defer s.restartMu.Unlock()
	s.killFFmpeg(true)
}

func (s *Session) killFFmpeg(wait bool) {
	if s.cmd != nil && s.cmd.Process != nil {
		// If the process was throttled (SIGSTOP), a pending SIGKILL from context
		// cancellation might not be processed until it receives SIGCONT.
		_ = s.cmd.Process.Signal(syscall.SIGCONT)
	}
	if s.cmdCancel != nil {
		s.cmdCancel()
		s.cmdCancel = nil
	}
	if wait && s.cmdExited != nil {
		<-s.cmdExited
	}
	s.cmd = nil
	s.cmdExited = nil
}

// SessionStartSegment returns the first segment number that the current ffmpeg
// process is encoding from.
func (s *Session) SessionStartSegment() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.startSegment
}

// GetExtinfSnapshot returns a copy of the EXTINF map for the given stream kind
// ("video", "audio0", etc.) accumulated so far by monitorJob.
func (s *Session) GetExtinfSnapshot(kind string) map[int]float64 {
	s.extinfMu.RLock()
	defer s.extinfMu.RUnlock()
	snap := make(map[int]float64)
	if m, ok := s.extinfMap[kind]; ok {
		for k, v := range m {
			snap[k] = v
		}
	}
	return snap
}

func (s *Session) ingestPlaylistFile(kind string) error {
	path := filepath.Join(s.OutputDir, hls.MediaPlaylistFilename(s.Prefix(), kind))
	b, err := os.ReadFile(path)
	if err != nil {
		return err // file might not exist yet
	}
	newMap, err := hls.ParseMediaPlaylist(b)
	if err != nil {
		return err
	}

	s.extinfMu.Lock()
	defer s.extinfMu.Unlock()
	if s.extinfMap == nil {
		s.extinfMap = make(map[string]map[int]float64)
	}
	if s.extinfMap[kind] == nil {
		s.extinfMap[kind] = make(map[int]float64)
	}
	hls.MergeExtinf(s.extinfMap[kind], newMap)
	return nil
}

func (s *Session) currentSegmentOnDisk(kind string) int {
	s.extinfMu.RLock()
	defer s.extinfMu.RUnlock()
	max := -1
	if m, ok := s.extinfMap[kind]; ok {
		for k := range m {
			if k > max {
				max = k
			}
		}
	}
	return max
}

// PrimeFromSegment ensures ffmpeg is running and positioned to decode segNum.
// Restarts ffmpeg (with a seek) when: not running, or the requested segment is
// behind the current start, or > 5 segments ahead of the encoded buffer.
func (s *Session) PrimeFromSegment(ctx context.Context, segNum int) error {
	s.Touch()

	s.restartMu.Lock()
	defer s.restartMu.Unlock()

	s.mu.Lock()
	if segNum > s.lastRequested {
		s.lastRequested = segNum
	}

	needRestart := false
	if s.cmd == nil {
		needRestart = true
	} else if segNum < s.startSegment {
		needRestart = true
	} else {
		// Use the max of the actual disk position or the start_segment-1.
		// If ffmpeg just started, disk position is -1, but it is working on startSegment.
		currentPos := s.currentSegmentOnDisk("video")
		if s.startSegment-1 > currentPos {
			currentPos = s.startSegment - 1
		}
		if segNum > currentPos+5 {
			needRestart = true
		}
	}
	s.mu.Unlock()

	if needRestart {
		s.log.Info("Restarting FFmpeg to seek timeline", "from_seg", segNum)
		s.killFFmpeg(true)

		s.mu.Lock()
		defer s.mu.Unlock()

		files, _ := filepath.Glob(filepath.Join(s.OutputDir, s.Prefix()+"*"))
		for _, f := range files {
			os.Remove(f)
		}

		s.extinfMu.Lock()
		s.extinfMap = make(map[string]map[int]float64)
		s.extinfMu.Unlock()

		if err := s.startFFmpegFrom(segNum); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) startFFmpegFrom(segNum int) error {
	s.startSegment = segNum

	cmdCtx, cancel := context.WithCancel(context.Background())
	s.cmdCancel = cancel

	opts := V2EncoderOpts{
		InputPath:         s.InputPath,
		OutputDir:         s.OutputDir,
		Prefix:            s.Prefix(),
		StartSegNum:       segNum,
		AudioTracks:       s.AudioTracks,
		VideoCopy:         s.Key.VideoCopy,
		SubtitleStreamIdx: s.Key.SubtitleStreamIdx,
		MaxHeight:         s.Key.MaxHeight,
		HwAccel:           s.HwAccel,
		SegLength:         s.SegLength,
	}

	s.log.Info("Starting V2 HLS Encoder", "start_seg", segNum, "video_copy", s.Key.VideoCopy)
	cmd, err := StartHLSV2Encoder(cmdCtx, opts)
	if err != nil {
		cancel()
		return err
	}
	s.cmd = cmd
	s.cmdExited = make(chan struct{})

	go func() {
		defer close(s.cmdExited)
		s.monitorJob(cmdCtx, cmd)
	}()
	return nil
}

func (s *Session) monitorJob(ctx context.Context, cmd *exec.Cmd) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	waitErrCh := make(chan error, 1)
	go func() { waitErrCh <- cmd.Wait() }()

	isThrottled := false

	for {
		select {
		case err := <-waitErrCh:
			exitCode := -1
			if cmd.ProcessState != nil {
				exitCode = cmd.ProcessState.ExitCode()
			}
			s.log.Warn("ffmpeg exited", "err", err, "exit_code", exitCode)
			s.mu.Lock()
			if s.cmd == cmd {
				s.cmd = nil
				if s.cmdCancel != nil {
					s.cmdCancel()
					s.cmdCancel = nil
				}
			}
			s.mu.Unlock()
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ingestPlaylistFile("video"); err != nil && !os.IsNotExist(err) {
				s.log.Warn("ingest video playlist failed", "err", err)
			}
			for i := range s.AudioTracks {
				_ = s.ingestPlaylistFile(fmt.Sprintf("audio%d", i))
			}

			s.mu.Lock()
			currentMax := s.currentSegmentOnDisk("video")
			buffer := currentMax - s.lastRequested

			// Throttle process if transcode is far ahead of player.
			if buffer > 20 && !isThrottled {
				s.log.Debug("Throttling ffmpeg (SIGSTOP)", "buffer", buffer)
				if err := cmd.Process.Signal(syscall.SIGSTOP); err == nil {
					isThrottled = true
				}
			} else if buffer < 10 && isThrottled {
				s.log.Debug("Resuming ffmpeg (SIGCONT)", "buffer", buffer)
				if err := cmd.Process.Signal(syscall.SIGCONT); err == nil {
					isThrottled = false
				}
			}
			s.mu.Unlock()
		}
	}
}

// RequestSegment ensures ffmpeg is positioned for segNum and blocks until the
// .m4s segment file appears on disk (or a 90-second timeout expires).
func (s *Session) RequestSegment(ctx context.Context, kind string, segNum int) error {
	s.Touch()

	s.mu.Lock()
	if segNum > s.lastRequested {
		s.lastRequested = segNum
	}
	s.mu.Unlock()

	segPath := filepath.Join(s.OutputDir, hls.SegmentFilename(s.Prefix(), kind, segNum))

	if err := s.PrimeFromSegment(ctx, segNum); err != nil {
		return err
	}

	return s.waitForFileOnDisk(ctx, segPath, 90*time.Second)
}

// RequestInitSegment blocks until the fMP4 init segment (.mp4) is available.
func (s *Session) RequestInitSegment(ctx context.Context, kind string) error {
	s.Touch()
	segPath := filepath.Join(s.OutputDir, hls.InitFilename(s.Prefix(), kind))
	return s.waitForFileOnDisk(ctx, segPath, 30*time.Second)
}

func (s *Session) waitForFileOnDisk(ctx context.Context, path string, timeout time.Duration) error {
	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for %s", filepath.Base(path))
		case <-ticker.C:
		}
	}
}
