package playback

import (
	"log"
	"os/exec"
	"sync/atomic"

	"github.com/thawng/velox/pkg/ffmpegbin"
)

var (
	vaapiTonemapOK  atomic.Bool
	vulkanPlaceboOK atomic.Bool
)

// ProbeTonemapCapabilitiesAsync initiates an eager detection in the background
// to check if the underlying GPU drivers actually support the zero-copy HDR filter chains.
func ProbeTonemapCapabilitiesAsync() {
	go func() {
		// 1. Probe VAAPI Tonemap (Intel/AMD)
		// Real-world resolution + h264_vaapi encode — many Intel iGPU drivers pass
		// a `-f null` smoke test but fail when ffmpeg auto-inserts a scaler between
		// tonemap_vaapi and h264_vaapi due to hw_frames_ctx mismatch (auto_scale_0
		// "Impossible to convert between the formats supported by..."). Encoding to
		// /dev/null catches that failure mode that the trivial 64x64+null path missed.
		probeOut := "/dev/null"
		cmd1 := exec.Command(ffmpegbin.FFmpeg(),
			"-hide_banner", "-loglevel", "error",
			"-vaapi_device", "/dev/dri/renderD128", // MUST be before -i so hwupload can initialize
			"-f", "lavfi", "-i", "color=red:s=1920x1080:d=0.1",
			"-vf", "format=p010,hwupload,tonemap_vaapi=format=nv12:matrix=bt709:primaries=bt709:transfer=bt709",
			"-c:v", "h264_vaapi", "-profile:v", "main", "-qp", "23",
			"-t", "0.1", "-f", "mp4", "-y", probeOut,
		)
		if err := cmd1.Run(); err == nil {
			vaapiTonemapOK.Store(true)
			log.Println("[INFO] HW Detect: VAAPI HDR tonemapping is supported by current driver.")
		} else {
			log.Printf("[INFO] HW Detect: VAAPI HDR tonemapping unavailable (%v) — falling back to software tonemapx.", err)
		}

		// 2. Probe Vulkan Libplacebo (NVIDIA/AMD)
		cmd2 := exec.Command(ffmpegbin.FFmpeg(),
			"-hide_banner", "-loglevel", "error",
			"-init_hw_device", "vulkan=vk:0",
			"-filter_hw_device", "vk",
			"-f", "lavfi", "-i", "color=red:s=64x64:d=0.1",
			"-vf", "format=p010,hwupload,libplacebo=format=yuv420p:colorspace=bt709:color_primaries=bt709:color_trc=bt709:tonemapping=mobius",
			"-f", "null", "-",
		)
		if err := cmd2.Run(); err == nil {
			vulkanPlaceboOK.Store(true)
			log.Println("[INFO] HW Detect: Vulkan libplacebo HDR tonemapping is supported by current driver.")
		}
	}()
}

// IsVAAPITonemapAvailable checks if eager probe found VAAPI Tonemap support
func IsVAAPITonemapAvailable() bool {
	return vaapiTonemapOK.Load()
}

// IsVulkanPlaceboAvailable checks if eager probe found Vulkan libplacebo support
func IsVulkanPlaceboAvailable() bool {
	return vulkanPlaceboOK.Load()
}
