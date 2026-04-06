package playback

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// hwAccelPriority defines the preference order for hardware accelerators.
var hwAccelPriority = []string{"videotoolbox", "nvenc", "vaapi", "qsv", "amf"}

type ffmpegCapabilities struct {
	hwaccels string
	encoders string
}

// DetectHWAccel probes FFmpeg for available hardware accelerators and returns
// the best available one. Returns "" if none found or FFmpeg is unavailable.
func DetectHWAccel() string {
	caps, ok := detectFFmpegCapabilities()
	if !ok {
		return ""
	}

	for _, accel := range hwAccelPriority {
		if hwAccelSupported(accel, caps) && hwAccelDeviceAvailable(accel) && hwAccelProbeOK(accel) {
			return accel
		}
	}
	return ""
}

// IsHWAccelAvailable reports whether a specific hardware accelerator is usable
// on the current host according to FFmpeg capabilities, device presence, and
// any runtime probe required for that backend.
func IsHWAccelAvailable(accel string) bool {
	if accel == "" || accel == "none" || accel == "auto" {
		return false
	}

	caps, ok := detectFFmpegCapabilities()
	if !ok {
		return false
	}

	return hwAccelSupported(accel, caps) && hwAccelDeviceAvailable(accel) && hwAccelProbeOK(accel)
}

func detectFFmpegCapabilities() (ffmpegCapabilities, bool) {
	hwaccels, ok := runFFmpegProbe("-hide_banner", "-hwaccels")
	if !ok {
		return ffmpegCapabilities{}, false
	}
	encoders, ok := runFFmpegProbe("-hide_banner", "-encoders")
	if !ok {
		return ffmpegCapabilities{}, false
	}
	return ffmpegCapabilities{
		hwaccels: hwaccels,
		encoders: encoders,
	}, true
}

func runFFmpegProbe(args ...string) (string, bool) {
	cmd := exec.Command("ffmpeg", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", false
	}
	return strings.ToLower(out.String()), true
}

func hwAccelSupported(accel string, caps ffmpegCapabilities) bool {
	switch accel {
	case "videotoolbox":
		return strings.Contains(caps.hwaccels, "videotoolbox") && strings.Contains(caps.encoders, "h264_videotoolbox")
	case "vaapi":
		return strings.Contains(caps.hwaccels, "vaapi") && strings.Contains(caps.encoders, "h264_vaapi")
	case "nvenc":
		return strings.Contains(caps.encoders, "h264_nvenc")
	case "qsv":
		return strings.Contains(caps.hwaccels, "qsv") && strings.Contains(caps.encoders, "h264_qsv")
	case "amf":
		return strings.Contains(caps.encoders, "h264_amf")
	default:
		return false
	}
}

// hwAccelDeviceAvailable checks whether the underlying device for a hardware
// accelerator actually exists at runtime. Prevents attempting VAAPI/NVENC on
// systems where the driver/device is absent (e.g. macOS host via OrbStack).
func hwAccelDeviceAvailable(accel string) bool {
	switch accel {
	case "videotoolbox":
		return runtime.GOOS == "darwin"
	case "vaapi":
		return hasDRMRenderNode()
	case "nvenc":
		if runtime.GOOS == "windows" {
			return true
		}
		_, err := os.Stat("/dev/nvidia0")
		return err == nil
	case "qsv":
		// QSV uses the same DRM render node as VAAPI on Linux.
		if runtime.GOOS == "windows" {
			return true
		}
		return hasDRMRenderNode()
	case "amf":
		// AMF is supported natively on Windows and can also work on Linux with
		// a custom FFmpeg + AMD driver stack. The encode probe below confirms it.
		return runtime.GOOS == "windows" || runtime.GOOS == "linux"
	default:
		return true
	}
}

func hasDRMRenderNode() bool {
	for _, dev := range []string{"/dev/dri/renderD128", "/dev/dri/renderD129", "/dev/dri/renderD130"} {
		if _, err := os.Stat(dev); err == nil {
			return true
		}
	}
	return false
}

func hwAccelProbeOK(accel string) bool {
	switch accel {
	case "vaapi":
		return vaAPIEncodeProbe()
	case "amf":
		return amfEncodeProbe()
	default:
		return true
	}
}

// vaAPIEncodeProbe runs a tiny FFmpeg encode to verify that H.264 VAAPI
// encoding actually works on this hardware/driver combination.
// Returns false if the driver reports "No usable encoding profile found"
// or any other encoder initialisation error.
func vaAPIEncodeProbe() bool {
	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "color=black:s=64x64:d=0.1",
		"-vf", "format=nv12,hwupload",
		"-vaapi_device", "/dev/dri/renderD128",
		"-c:v", "h264_vaapi",
		"-frames:v", "1",
		"-f", "null", "-",
	)
	return cmd.Run() == nil
}

// amfEncodeProbe runs a tiny FFmpeg encode to verify that H.264 AMF is usable.
// AMF is encoder-based, so we only need to prove that FFmpeg can initialize
// the encoder successfully on this host.
func amfEncodeProbe() bool {
	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "color=black:s=64x64:d=0.1",
		"-c:v", "h264_amf",
		"-frames:v", "1",
		"-f", "null", "-",
	)
	return cmd.Run() == nil
}
