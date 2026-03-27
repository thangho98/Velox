package playback

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// hwAccelPriority defines the preference order for hardware accelerators.
var hwAccelPriority = []string{"videotoolbox", "nvenc", "vaapi", "qsv"}

// DetectHWAccel probes FFmpeg for available hardware accelerators and returns
// the best available one. Returns "" if none found or FFmpeg is unavailable.
func DetectHWAccel() string {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-hwaccels")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}

	lower := strings.ToLower(out.String())
	for _, accel := range hwAccelPriority {
		if strings.Contains(lower, accel) && hwAccelDeviceAvailable(accel) {
			return accel
		}
	}
	return ""
}

// hwAccelDeviceAvailable checks whether the underlying device for a hardware
// accelerator actually exists at runtime. Prevents attempting VAAPI/NVENC on
// systems where the driver/device is absent (e.g. macOS host via OrbStack).
func hwAccelDeviceAvailable(accel string) bool {
	switch accel {
	case "videotoolbox":
		return runtime.GOOS == "darwin"
	case "vaapi":
		// Require at least one DRM render node
		devFound := false
		for _, dev := range []string{"/dev/dri/renderD128", "/dev/dri/renderD129", "/dev/dri/renderD130"} {
			if _, err := os.Stat(dev); err == nil {
				devFound = true
				break
			}
		}
		if !devFound {
			return false
		}
		// Verify the driver can actually encode H.264 by running a tiny test encode.
		return vaAPIEncodeProbe()
	case "nvenc":
		_, err := os.Stat("/dev/nvidia0")
		return err == nil
	case "qsv":
		// QSV uses the same DRM render node as VAAPI on Linux
		_, err := os.Stat("/dev/dri/renderD128")
		return err == nil
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
