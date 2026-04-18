package handler

import (
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/pkg/ffprobe"
)

// resolveIsHDR returns whether a media file is HDR.
// For cloud files, uses persisted DB fields (ffprobe can't access fshare:// paths).
// For local files, runs ffprobe for accurate real-time detection.
func resolveIsHDR(mf *model.MediaFile) bool {
	if scanner.IsCloudPath(mf.FilePath) {
		return isHDRFromDB(mf)
	}
	return ffprobe.IsHDRLike(mf.FilePath)
}

// resolveNeedsTonemap returns whether server-side tone mapping is needed.
// For cloud files, uses persisted DB fields to determine if color metadata is missing.
// For local files, runs ffprobe for accurate detection.
func resolveNeedsTonemap(mf *model.MediaFile) bool {
	if scanner.IsCloudPath(mf.FilePath) {
		if !isHDRFromDB(mf) {
			return false
		}
		// HDR but missing color metadata → needs server tonemap
		return isUnknownOrEmpty(mf.ColorTransfer) ||
			isUnknownOrEmpty(mf.ColorPrimaries)
	}
	return ffprobe.NeedsHDRColorMetadataFallback(mf.FilePath)
}

func isHDRFromDB(mf *model.MediaFile) bool {
	if mf.IsHDR || mf.DVProfile > 0 {
		return true
	}
	ct := strings.ToLower(mf.ColorTransfer)
	cp := strings.ToLower(mf.ColorPrimaries)
	return strings.Contains(ct, "smpte2084") ||
		strings.Contains(ct, "arib-std-b67") ||
		strings.Contains(cp, "bt2020")
}

func isUnknownOrEmpty(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "" || v == "unknown" || v == "reserved"
}
