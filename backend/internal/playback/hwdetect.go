package playback

import (
	"bytes"
	"os/exec"
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
		if strings.Contains(lower, accel) {
			return accel
		}
	}
	return ""
}
