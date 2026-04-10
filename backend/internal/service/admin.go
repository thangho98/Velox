package service

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/ffmpegbin"
)

// AdminService provides admin dashboard functionality.
type AdminService struct {
	adminRepo     *repository.AdminRepo
	libraryRepo   *repository.LibraryRepo
	mediaFileRepo *repository.MediaFileRepo
	userRepo      *repository.UserRepo
	startTime     time.Time
	hwAccel       string
	dbPath        string
}

func NewAdminService(
	db repository.DBTX,
	libraryRepo *repository.LibraryRepo,
	mediaFileRepo *repository.MediaFileRepo,
	userRepo *repository.UserRepo,
	startTime time.Time,
	hwAccel string,
	dbPath string,
) *AdminService {
	return &AdminService{
		adminRepo:     repository.NewAdminRepo(db),
		libraryRepo:   libraryRepo,
		mediaFileRepo: mediaFileRepo,
		userRepo:      userRepo,
		startTime:     startTime,
		hwAccel:       hwAccel,
		dbPath:        dbPath,
	}
}

// GetServerInfo returns server status information.
func (s *AdminService) GetServerInfo(ctx context.Context) (*model.ServerInfo, error) {
	info := &model.ServerInfo{
		Version:   "0.1.0",
		Uptime:    time.Since(s.startTime).Truncate(time.Second).String(),
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Database:  s.dbPath,
		HWAccel:   s.hwAccel,
	}

	if info.HWAccel == "" {
		info.HWAccel = "none"
	}

	// FFmpeg version
	info.FFmpegVer = detectFFmpegVersion()

	// Counts from DB
	var err error
	info.MediaCount, err = s.adminRepo.CountTable(ctx, "media")
	if err != nil {
		return nil, fmt.Errorf("counting media: %w", err)
	}

	info.SeriesCount, err = s.adminRepo.CountTable(ctx, "series")
	if err != nil {
		return nil, fmt.Errorf("counting series: %w", err)
	}

	info.UserCount, err = s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}

	// Total file size
	info.TotalSize, err = s.mediaFileRepo.TotalSize(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting total size: %w", err)
	}

	return info, nil
}

// GetLibraryStats returns per-library statistics.
func (s *AdminService) GetLibraryStats(ctx context.Context) ([]model.LibraryStats, error) {
	return s.libraryRepo.GetStats(ctx)
}

func detectFFmpegVersion() string {
	out, err := exec.Command(ffmpegbin.FFmpeg(), "-version").Output()
	if err != nil {
		return "not found"
	}
	// First line is "ffmpeg version X.Y.Z ..."
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			return parts[2]
		}
		return lines[0]
	}
	return "unknown"
}
