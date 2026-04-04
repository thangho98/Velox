package transcoder

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// hasSubtitlesFilter is set once at init — true when FFmpeg was built with libass.
var hasSubtitlesFilter = detectSubtitlesFilter()

// SupportsSubtitleBurnIn reports whether the local FFmpeg build can burn subtitles.
func SupportsSubtitleBurnIn() bool {
	return hasSubtitlesFilter
}

func detectSubtitlesFilter() bool {
	out, err := exec.Command("ffmpeg", "-filters").CombinedOutput()
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
		return []string{"-hwaccel", "amf"}
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

func buildFFmpegInputArgs(hwAccel string) []string {
	args := ffmpegInputProbeArgs()
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

// hdrTonemapVAAPI returns VAAPI-native HDR→SDR tone mapping filter.
// Uses VAAPI's built-in tonemap which runs on GPU — fast and efficient.
// Output is NV12 ready for h264_vaapi encoder.
func hdrTonemapVAAPI() string {
	return "tonemap_vaapi=format=nv12"
}

// hdrTonemapOpenCL returns OpenCL-based HDR→SDR tone mapping filter.
// Works with NVIDIA/AMD GPUs via OpenCL. Output is NV12.
func hdrTonemapOpenCL() string {
	return "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=opencl:format=nv12"
}

// hdrToneMapFilterForHW returns the appropriate HDR→SDR tone mapping filter
// based on the hardware accelerator in use. Falls back to software zscale+hable
// when no hardware accelerator is available or for unsupported cases.
func hdrToneMapFilterForHW(hwAccel string) string {
	switch hwAccel {
	case "vaapi":
		return hdrTonemapVAAPI()
	case "nvenc", "amf":
		return hdrTonemapOpenCL()
	default:
		return hdrToneMapFilter() // software fallback: zscale+hable
	}
}

// buildVideoEncodeArgs builds -vf + -c:v args for single-quality HLS encoding.
// Handles HW encoder selection, HDR→SDR tone mapping, and subtitle burn-in.
func buildVideoEncodeArgs(hwAccel string, hdr bool, siIdx int, inputPath string) []string {
	var filters []string
	if hdr {
		filters = append(filters, hdrToneMapFilterForHW(hwAccel))
	}
	if siIdx >= 0 && hasSubtitlesFilter {
		escaped := escapeFFmpegSubtitlePath(inputPath)
		filters = append(filters, fmt.Sprintf("subtitles=filename='%s':si=%d", escaped, siIdx))
	}
	// VAAPI: force NV12 surface format so h264_vaapi can encode 10-bit sources.
	// When HDR tonemap is applied via tonemap_vaapi, format is already NV12.
	if hwAccel == "vaapi" && len(filters) == 0 {
		filters = append(filters, "scale_vaapi=format=nv12")
	}

	var args []string
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args, "-c:v", hwVideoCodec(hwAccel))
	args = append(args, hwEncoderArgs(hwAccel)...)
	return args
}

// buildImageSubtitleBurnInArgs burns a bitmap subtitle stream (PGS/VobSub) into
// the primary video using filter_complex overlay. The selected subtitle stream is
// referenced by absolute stream index on input 0.
func buildImageSubtitleBurnInArgs(hwAccel string, hdr bool, subtitleStreamIndex int) []string {
	complexFilter := buildImageSubtitleBurnInFilter(hwAccel, hdr, subtitleStreamIndex)
	args := []string{
		"-filter_complex", complexFilter,
		"-map", "[vout]",
		"-map", "0:a:0?",
		"-c:v", hwVideoCodec(hwAccel),
	}
	args = append(args, hwEncoderArgs(hwAccel)...)
	return args
}

// buildImageSubtitleBurnInVideoOnlyArgs is the same burn-in path as
// buildImageSubtitleBurnInArgs, but only maps the filtered video output.
func buildImageSubtitleBurnInVideoOnlyArgs(hwAccel string, hdr bool, subtitleStreamIndex int) []string {
	complexFilter := buildImageSubtitleBurnInFilter(hwAccel, hdr, subtitleStreamIndex)
	args := []string{
		"-filter_complex", complexFilter,
		"-map", "[vout]",
		"-c:v", hwVideoCodec(hwAccel),
	}
	args = append(args, hwEncoderArgs(hwAccel)...)
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
// For HDR sources, uses zscale+hable tonemap (software) because subtitle overlay
// via filter_complex requires system memory frames. VAAPI hardware tonemap outputs
// to GPU surfaces which can't be directly overlaid with subtitles.
func buildImageSubtitleBurnInFilter(hwAccel string, hdr bool, subtitleStreamIndex int) string {
	if hdr {
		// Always use zscale+hable for HDR subtitle burn-in — filter_complex
		// requires software processing, so VAAPI hwaccel doesn't help here.
		return fmt.Sprintf("[0:v:0]%s[base];[base][0:%d]overlay[vout]", hdrToneMapFilter(), subtitleStreamIndex)
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
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "v:0",
		"-show_entries", "stream=color_transfer,color_primaries",
		"-of", "default=noprint_wrappers=1",
		inputPath,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("transcoder: HDR probe failed for %q: %v", inputPath, err)
		return false
	}
	lower := strings.ToLower(out.String())
	return strings.Contains(lower, "smpte2084") || strings.Contains(lower, "bt2020")
}

// hdrToneMapFilter returns the FFmpeg -vf filter chain for HDR→SDR tone mapping.
// Output is SDR BT.709 in yuv420p, suitable for H.264/H.265 HLS streaming.
func hdrToneMapFilter() string {
	return "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable,zscale=t=bt709:m=bt709,format=yuv420p"
}
