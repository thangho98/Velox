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
		// Requires standard vaapi format upconversion (p010 -> nv12) AND device init BEFORE filters.
		cmd1 := exec.Command(ffmpegbin.FFmpeg(),
			"-hide_banner", "-loglevel", "error",
			"-vaapi_device", "/dev/dri/renderD128", // MUST be before -i so hwupload can initialize
			"-f", "lavfi", "-i", "color=red:s=64x64:d=0.1",
			"-vf", "format=p010,hwupload,tonemap_vaapi=format=nv12:matrix=bt709:primaries=bt709:transfer=bt709",
			"-f", "null", "-",
		)
		if err := cmd1.Run(); err == nil {
			vaapiTonemapOK.Store(true)
			log.Println("[INFO] HW Detect: VAAPI HDR tonemapping is supported by current driver.")
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
