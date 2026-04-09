package hls

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var (
	rxMediaPlaylist = regexp.MustCompile(`^v2_ss([a-zA-Z0-9-]+)_(\d+)_f(\d+)_si(-?\d+)_vc([01])_h(\d+)_([a-zA-Z0-9]+)\.m3u8$`)
	rxMediaSegment  = regexp.MustCompile(`^v2_ss([a-zA-Z0-9-]+)_(\d+)_f(\d+)_si(-?\d+)_vc([01])_h(\d+)_([a-zA-Z0-9]+)_(-?\d+)\.(m4s|mp4)$`)
)

// SessionKey uniquely identifies an HLS transcode session
type SessionKey struct {
	StreamSessionID   string
	MediaID           int64
	FileID            int64
	SubtitleStreamIdx int
	VideoCopy         bool
	MaxHeight         int
}

func BuildPrefix(key SessionKey) string {
	videoCopyInt := 0
	if key.VideoCopy {
		videoCopyInt = 1
	}
	return fmt.Sprintf("v2_ss%s_%d_f%d_si%d_vc%d_h%d_",
		key.StreamSessionID, key.MediaID, key.FileID, key.SubtitleStreamIdx, videoCopyInt, key.MaxHeight)
}

func SegmentNumber(time, segLength float64) int {
	return int(math.Floor(time / segLength))
}

func SegmentTimeRange(num int, segLength float64) (start, end float64) {
	return float64(num) * segLength, float64(num+1) * segLength
}

func SegmentFilename(prefix, kind string, idx int) string {
	return fmt.Sprintf("%s%s_%d.m4s", prefix, kind, idx)
}

func InitFilename(prefix, kind string) string {
	return fmt.Sprintf("%s%s_-1.mp4", prefix, kind)
}

func MediaPlaylistFilename(prefix, kind string) string {
	return fmt.Sprintf("%s%s.m3u8", prefix, kind)
}

// ParseFilename parses our v2 prefix structure.
// Returns SessionKey, kind (e.g. video, audio0), segNum (if .m4s or .mp4, -1 for .mp4), and error.
func ParseFilename(filename string) (SessionKey, string, int, error) {
	if m := rxMediaPlaylist.FindStringSubmatch(filename); m != nil {
		mediaID, _ := strconv.ParseInt(m[2], 10, 64)
		fileID, _ := strconv.ParseInt(m[3], 10, 64)
		si, _ := strconv.Atoi(m[4])
		vc, _ := strconv.Atoi(m[5])
		h, _ := strconv.Atoi(m[6])
		return SessionKey{
			StreamSessionID:   m[1],
			MediaID:           mediaID,
			FileID:            fileID,
			SubtitleStreamIdx: si,
			VideoCopy:         vc == 1,
			MaxHeight:         h,
		}, m[7], 0, nil
	}

	if m := rxMediaSegment.FindStringSubmatch(filename); m != nil {
		mediaID, _ := strconv.ParseInt(m[2], 10, 64)
		fileID, _ := strconv.ParseInt(m[3], 10, 64)
		si, _ := strconv.Atoi(m[4])
		vc, _ := strconv.Atoi(m[5])
		h, _ := strconv.Atoi(m[6])
		num, _ := strconv.Atoi(m[8])
		return SessionKey{
			StreamSessionID:   m[1],
			MediaID:           mediaID,
			FileID:            fileID,
			SubtitleStreamIdx: si,
			VideoCopy:         vc == 1,
			MaxHeight:         h,
		}, m[7], num, nil
	}

	return SessionKey{}, "", 0, fmt.Errorf("invalid filename format: %s", filename)
}
