package transcoder

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/thawng/velox/pkg/ffmpegbin"
	"github.com/thawng/velox/pkg/ffprobe"
)

// hasSubtitlesFilter is set once at init — true when FFmpeg was built with libass.
var hasSubtitlesFilter = detectSubtitlesFilter()

// SupportsSubtitleBurnIn reports whether the local FFmpeg build can burn subtitles.
func SupportsSubtitleBurnIn() bool {
	return hasSubtitlesFilter
}

func detectSubtitlesFilter() bool {
	out, err := exec.Command(ffmpegbin.FFmpeg(), "-filters").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "subtitles")
}

// hwInputArgs returns FFmpeg input-side args for the given hardware accelerator.
func hwInputArgs(hwAccel string) []string {
	switch hwAccel {
	case "videotoolbox":
		return []string{"-hwaccel", "videotoolbox"}
	case "vaapi":
		return []string{"-hwaccel", "vaapi", "-hwaccel_output_format", "vaapi", "-hwaccel_device", "/dev/dri/renderD128"}
	case "nvenc":
		return []string{"-hwaccel", "cuda"}
	case "qsv":
		return []string{"-hwaccel", "qsv"}
	case "amf":
		// AMF is an encoder backend, not a generic FFmpeg hwaccel name.
		// Keep decode in software unless a platform-specific path is added later.
		return nil
	}
	return nil
}

// hwScaleFilter returns the appropriate scale filter for the given HW accelerator.
// For VAAPI, forces NV12 output format so h264_vaapi can encode 10-bit sources
// (e.g. HEVC P010) that would otherwise fail with "No usable encoding profile".
func hwScaleFilter(hwAccel string, height int) string {
	switch hwAccel {
	case "vaapi":
		return fmt.Sprintf("scale_vaapi=w=-2:h=%d:format=nv12", height)
	default:
		return fmt.Sprintf("scale=-2:%d", height)
	}
}

// ffmpegInputProbeArgs increases demux probing so image-based subtitle streams
// like PGS are discovered reliably before we attempt burn-in.
func ffmpegInputProbeArgs() []string {
	return []string{
		"-probesize", "50000000",
		"-analyzeduration", "100000000",
	}
}

// buildFFmpegInputArgs returns input-side FFmpeg args (probe + hwaccel).
// For HDR: always uses software decode so CPU-decoded frames retain correct
// color metadata for the tonemapx filter chain. Hardware decoders
// (VideoToolbox, QSV, AMF) can strip or alter color metadata on the decoded
// frames, causing tonemapx to misinterpret the color space.
func buildFFmpegInputArgs(hwAccel string, hdr bool) []string {
	args := ffmpegInputProbeArgs()
	if hdr {
		// Software decode for ALL HDR — tonemapx needs correct color metadata.
		if hwAccel == "vaapi" {
			// VAAPI still needs the device init for hwupload → h264_vaapi encode.
			args = append(args, "-vaapi_device", "/dev/dri/renderD128")
		}
		// All other hwAccel types (nvenc, videotoolbox, qsv, amf):
		// skip hwaccel decode args — decode in software, encode with HW encoder.
		return args
	}
	args = append(args, hwInputArgs(hwAccel)...)
	return args
}

// hwVideoCodec returns the FFmpeg video encoder for the given HW accelerator.
// Falls back to libx264 when hwAccel is empty.
func hwVideoCodec(hwAccel string) string {
	switch hwAccel {
	case "videotoolbox":
		return "h264_videotoolbox"
	case "vaapi":
		return "h264_vaapi"
	case "nvenc":
		return "h264_nvenc"
	case "qsv":
		return "h264_qsv"
	case "amf":
		return "h264_amf"
	}
	return "libx264"
}

// hdrToneMapFilterSW returns the software HDR→SDR tonemap filter chain.
// Uses tonemapx (Jellyfin SIMD-optimized) which correctly handles both standard
// HDR10 (BT.2020/PQ) and Dolby Vision (including Profile 5 IPT-PQ-C2).
// The older zscale chain fails on DV Profile 5 because it assumes BT.2020nc
// matrix coefficients, but DV P5 uses IPT color space — producing green/magenta tint.
// outFmt controls the final pixel format: "yuv420p" for software encode,
// "nv12" for VAAPI hwupload.
func hdrToneMapFilterSW(inputPath, outFmt string) string {
	return hdrColorMetadataFallbackPrefix(inputPath) +
		fmt.Sprintf(
			"tonemapx=tonemap=bt2390:desat=0:peak=100:t=bt709:m=bt709:p=bt709,format=%s",
			outFmt,
		)
}

// hdrToneMapFilterForHW returns the appropriate HDR→SDR tone mapping filter
// for the given hardware accelerator.
// Always uses software tonemapx (CPU SIMD) because:
// 1. tonemap_vaapi does not reliably convert BT.2020 primaries on many GPU drivers
// 2. tonemapx correctly handles Dolby Vision (including DV Profile 5)
func hdrToneMapFilterForHW(hwAccel, inputPath string) string {
	switch hwAccel {
	case "vaapi":
		// Software tonemap → NV12 → hwupload to VAAPI surface for h264_vaapi.
		return hdrToneMapFilterSW(inputPath, "nv12") + ",hwupload"
	default:
		return hdrToneMapFilter(inputPath) // software: tonemapx → yuv420p
	}
}

// buildVideoEncodeArgs builds -vf + -c:v args for single-quality HLS encoding.
// Handles HW encoder selection, HDR→SDR tone mapping, and subtitle burn-in.
func buildVideoEncodeArgs(hwAccel string, hdr bool, siIdx int, inputPath string) []string {
	var filters []string
	if hdr {
		filters = append(filters, hdrToneMapFilterForHW(hwAccel, inputPath))
	}
	if siIdx >= 0 && hasSubtitlesFilter {
		escaped := escapeFFmpegSubtitlePath(inputPath)
		filters = append(filters, fmt.Sprintf("subtitles=filename='%s':si=%d", escaped, siIdx))
	}
	// VAAPI: force NV12 surface format so h264_vaapi can encode 10-bit sources.
	// When HDR tonemap is applied, format is already NV12+hwupload.
	if hwAccel == "vaapi" && len(filters) == 0 {
		filters = append(filters, "scale_vaapi=format=nv12")
	}

	var args []string
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args, "-c:v", hwVideoCodec(hwAccel))
	args = append(args, hwEncoderArgs(hwAccel)...)

	// Tag the output stream as BT.709 so players render correct colors.
	if hdr {
		args = append(args, "-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709")
	}
	return args
}

// buildImageSubtitleBurnInArgs burns a bitmap subtitle stream (PGS/VobSub) into
// the primary video using filter_complex overlay. The selected subtitle stream is
// referenced by absolute stream index on input 0.
func buildImageSubtitleBurnInArgs(hwAccel string, hdr bool, inputPath string, subtitleStreamIndex int) []string {
	complexFilter := buildImageSubtitleBurnInFilter(hwAccel, hdr, inputPath, subtitleStreamIndex)
	args := []string{
		"-filter_complex", complexFilter,
		"-map", "[vout]",
		"-map", "0:a:0?",
		"-c:v", hwVideoCodec(hwAccel),
	}
	args = append(args, hwEncoderArgs(hwAccel)...)
	if hdr {
		args = append(args, "-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709")
	}
	return args
}

// buildImageSubtitleBurnInVideoOnlyArgs is the same burn-in path as
// buildImageSubtitleBurnInArgs, but only maps the filtered video output.
func buildImageSubtitleBurnInVideoOnlyArgs(hwAccel string, hdr bool, inputPath string, subtitleStreamIndex int) []string {
	complexFilter := buildImageSubtitleBurnInFilter(hwAccel, hdr, inputPath, subtitleStreamIndex)
	args := []string{
		"-filter_complex", complexFilter,
		"-map", "[vout]",
		"-c:v", hwVideoCodec(hwAccel),
	}
	args = append(args, hwEncoderArgs(hwAccel)...)
	if hdr {
		args = append(args, "-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709")
	}
	return args
}

// hwEncoderArgs returns encoder-specific args (profile, quality, pixel format)
// for the given HW accelerator. Centralizes encoder tuning so all encode paths
// use consistent settings.
func hwEncoderArgs(hwAccel string) []string {
	switch hwAccel {
	case "":
		return []string{"-preset", "veryfast", "-crf", "23", "-threads", "0", "-pix_fmt", "yuv420p"}
	case "vaapi":
		return []string{"-profile:v", "main", "-qp", "23"}
	case "amf":
		return []string{"-preset", "veryfast", "-quality", "balanced"}
	default:
		return nil
	}
}

// buildImageSubtitleBurnInFilter builds the filter_complex string for burning
// image-based subtitles (PGS/VobSub) into the video.
// For HDR sources, uses tonemapx (software) because subtitle overlay via
// filter_complex requires system memory frames.
func buildImageSubtitleBurnInFilter(hwAccel string, hdr bool, inputPath string, subtitleStreamIndex int) string {
	if hdr {
		// Always use software tonemap for HDR subtitle burn-in — filter_complex
		// requires software processing, so VAAPI hwaccel doesn't help here.
		return fmt.Sprintf("[0:v:0]%s[base];[base][0:%d]overlay[vout]", hdrToneMapFilter(inputPath), subtitleStreamIndex)
	}
	return fmt.Sprintf("[0:v:0][0:%d]overlay[vout]", subtitleStreamIndex)
}

// escapeFFmpegSubtitlePath escapes a file path for FFmpeg's subtitles filter.
// The subtitles filter uses libass which requires escaping at two levels:
//  1. Filter option level: : ; [ ] ' \
//  2. The path is NOT wrapped in quotes — all special chars are backslash-escaped.
func escapeFFmpegSubtitlePath(path string) string {
	// Order matters: escape backslash first
	path = strings.ReplaceAll(path, `\`, `\\`)
	path = strings.ReplaceAll(path, `'`, `\'`)
	path = strings.ReplaceAll(path, `:`, `\:`)
	path = strings.ReplaceAll(path, `[`, `\[`)
	path = strings.ReplaceAll(path, `]`, `\]`)
	path = strings.ReplaceAll(path, `;`, `\;`)
	return path
}

// --- HDR detection ---

// isHDRFile returns true if the input file's primary video stream uses HDR
// color transfer (PQ/SMPTE2084) or BT.2020 color primaries.
func isHDRFile(inputPath string) bool {
	return ffprobe.IsHDRLike(inputPath)
}

// hdrToneMapFilter returns the FFmpeg -vf filter chain for HDR→SDR tone mapping.
// Output is SDR BT.709 in yuv420p, suitable for H.264/H.265 HLS streaming.
func hdrToneMapFilter(inputPath string) string {
	return hdrToneMapFilterSW(inputPath, "yuv420p")
}

func hdrColorMetadataFallbackPrefix(inputPath string) string {
	if !ffprobe.NeedsHDRColorMetadataFallback(inputPath) {
		return ""
	}
	return "setparams=color_primaries=bt2020:color_trc=smpte2084:colorspace=bt2020nc,"
}
