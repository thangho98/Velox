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
)

// AdminService provides admin dashboard functionality.
type AdminService struct {
	db        repository.DBTX
	userRepo  *repository.UserRepo
	startTime time.Time
	hwAccel   string
	dbPath    string
}

func NewAdminService(db repository.DBTX, userRepo *repository.UserRepo, startTime time.Time, hwAccel, dbPath string) *AdminService {
	return &AdminService{
		db:        db,
		userRepo:  userRepo,
		startTime: startTime,
		hwAccel:   hwAccel,
		dbPath:    dbPath,
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
	info.MediaCount, err = s.countTable(ctx, "media")
	if err != nil {
		return nil, fmt.Errorf("counting media: %w", err)
	}

	info.SeriesCount, err = s.countTable(ctx, "series")
	if err != nil {
		return nil, fmt.Errorf("counting series: %w", err)
	}

	info.UserCount, err = s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}

	// Total file size
	row := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(file_size), 0) FROM media_files`)
	if err := row.Scan(&info.TotalSize); err != nil {
		return nil, fmt.Errorf("summing file sizes: %w", err)
	}

	return info, nil
}

// GetLibraryStats returns per-library statistics.
func (s *AdminService) GetLibraryStats(ctx context.Context) ([]model.LibraryStats, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			l.id, l.name, l.type,
			(SELECT COUNT(*) FROM media m WHERE m.library_id = l.id) +
			(SELECT COUNT(*) FROM series sr WHERE sr.library_id = l.id) as item_count,
			(SELECT COUNT(*) FROM media_files mf
			 JOIN media m2 ON mf.media_id = m2.id WHERE m2.library_id = l.id) as file_count,
			(SELECT COALESCE(SUM(mf2.file_size), 0) FROM media_files mf2
			 JOIN media m3 ON mf2.media_id = m3.id WHERE m3.library_id = l.id) as total_size,
			COALESCE((SELECT MAX(sj.finished_at) FROM scan_jobs sj
			          WHERE sj.library_id = l.id AND sj.status = 'completed'), '') as last_scanned
		FROM libraries l
		ORDER BY l.name`)
	if err != nil {
		return nil, fmt.Errorf("querying library stats: %w", err)
	}
	defer rows.Close()

	var stats []model.LibraryStats
	for rows.Next() {
		var ls model.LibraryStats
		if err := rows.Scan(&ls.ID, &ls.Name, &ls.Type, &ls.ItemCount, &ls.FileCount, &ls.TotalSize, &ls.LastScanned); err != nil {
			return nil, fmt.Errorf("scanning library stats: %w", err)
		}
		stats = append(stats, ls)
	}
	return stats, rows.Err()
}

func (s *AdminService) countTable(ctx context.Context, table string) (int, error) {
	var count int
	row := s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
	err := row.Scan(&count)
	return count, err
}

func detectFFmpegVersion() string {
	out, err := exec.Command("ffmpeg", "-version").Output()
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
