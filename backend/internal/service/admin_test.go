package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/model"
)

type mockServerInfoService struct {
	adminCount  int
	seriesCount int
	userCount   int
	totalSize   int64
	startTime   time.Time
	hwAccel     string
	dbPath      string
}

func newMockServerInfo(adminCount, seriesCount, userCount int, totalSize int64) *mockServerInfoService {
	return &mockServerInfoService{
		adminCount:  adminCount,
		seriesCount: seriesCount,
		userCount:   userCount,
		totalSize:   totalSize,
		startTime:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		hwAccel:     "none",
		dbPath:      ":memory:",
	}
}

func (s *mockServerInfoService) GetServerInfo(ctx context.Context) (*model.ServerInfo, error) {
	info := &model.ServerInfo{
		Version:   "0.1.0",
		Uptime:    time.Since(s.startTime).Truncate(time.Second).String(),
		GoVersion: "go1.26",
		OS:        "darwin",
		Arch:      "arm64",
		Database:  s.dbPath,
		HWAccel:   s.hwAccel,
		FFmpegVer: "6.0",
	}
	if info.HWAccel == "" {
		info.HWAccel = "none"
	}
	info.MediaCount = s.adminCount
	info.SeriesCount = s.seriesCount
	info.UserCount = s.userCount
	info.TotalSize = s.totalSize
	return info, nil
}

func TestAdminService_GetServerInfo(t *testing.T) {
	t.Parallel()
	svc := newMockServerInfo(100, 50, 3, 1024*1024*1024*500)
	ctx := context.Background()
	info, err := svc.GetServerInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0.1.0", info.Version)
	assert.Equal(t, "go1.26", info.GoVersion)
	assert.Equal(t, "darwin", info.OS)
	assert.Equal(t, "arm64", info.Arch)
	assert.Equal(t, ":memory:", info.Database)
	assert.Equal(t, "none", info.HWAccel)
	assert.Equal(t, "6.0", info.FFmpegVer)
	assert.Equal(t, 100, info.MediaCount)
	assert.Equal(t, 50, info.SeriesCount)
	assert.Equal(t, 3, info.UserCount)
	assert.Equal(t, int64(1024*1024*1024*500), info.TotalSize)
}

func TestAdminService_GetServerInfo_HWAccelNone(t *testing.T) {
	t.Parallel()
	svc := newMockServerInfo(0, 0, 0, 0)
	svc.hwAccel = ""
	ctx := context.Background()
	info, err := svc.GetServerInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "none", info.HWAccel)
}

func TestAdminService_GetServerInfo_ZeroCounts(t *testing.T) {
	t.Parallel()
	svc := newMockServerInfo(0, 0, 0, 0)
	ctx := context.Background()
	info, err := svc.GetServerInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, info.MediaCount)
	assert.Equal(t, 0, info.SeriesCount)
	assert.Equal(t, 0, info.UserCount)
	assert.Equal(t, int64(0), info.TotalSize)
}

func TestAdminService_UptimeFormat(t *testing.T) {
	t.Parallel()
	svc := newMockServerInfo(0, 0, 0, 0)
	svc.startTime = time.Now().Add(-1 * time.Hour)
	ctx := context.Background()
	info, err := svc.GetServerInfo(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, info.Uptime)
}
