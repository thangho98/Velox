package ffprobe

import "testing"

func TestIsHDRLikeFromDetailed_DolbyVisionSideData(t *testing.T) {
	detailed := DetailedProbeResult{
		Streams: []StreamInfo{
			{
				CodecType: "video",
				Profile:   "Main 10",
				SideDataList: []SideDataInfo{
					{SideDataType: "DOVI configuration record", DVProfile: 5},
				},
			},
		},
	}

	if !isHDRLikeFromDetailed(detailed, "/media/movie.mkv") {
		t.Fatal("expected Dolby Vision side data to be treated as HDR-like")
	}
}

func TestIsHDRLikeFromDetailed_FilenameFallback(t *testing.T) {
	detailed := DetailedProbeResult{
		Streams: []StreamInfo{
			{
				CodecType: "video",
				Profile:   "Main 10",
			},
		},
	}

	if !isHDRLikeFromDetailed(detailed, "/media/Avatar 2025 DV WEB-DL.mkv") {
		t.Fatal("expected DV filename hint to be treated as HDR-like")
	}
}

func TestIsHDRLikeFromDetailed_SDRFile(t *testing.T) {
	detailed := DetailedProbeResult{
		Streams: []StreamInfo{
			{
				CodecType:      "video",
				Profile:        "High",
				ColorTransfer:  "bt709",
				ColorPrimaries: "bt709",
				ColorSpace:     "bt709",
			},
		},
	}

	if isHDRLikeFromDetailed(detailed, "/media/regular-sdr.mkv") {
		t.Fatal("expected SDR file to remain SDR")
	}
}

func TestNeedsHDRColorMetadataFallbackFromDetailed_DolbyVisionUnknownColorTags(t *testing.T) {
	detailed := DetailedProbeResult{
		Streams: []StreamInfo{
			{
				CodecType:      "video",
				Profile:        "Main 10",
				ColorTransfer:  "unknown",
				ColorPrimaries: "unknown",
				ColorSpace:     "unknown",
				SideDataList: []SideDataInfo{
					{SideDataType: "DOVI configuration record", DVProfile: 5},
				},
			},
		},
	}

	if !needsHDRColorMetadataFallbackFromDetailed(detailed, "/media/movie.mkv") {
		t.Fatal("expected Dolby Vision source with unknown color tags to need fallback")
	}
}

func TestNeedsHDRColorMetadataFallbackFromDetailed_HDR10KnownColorTags(t *testing.T) {
	detailed := DetailedProbeResult{
		Streams: []StreamInfo{
			{
				CodecType:      "video",
				Profile:        "Main 10",
				ColorTransfer:  "smpte2084",
				ColorPrimaries: "bt2020",
				ColorSpace:     "bt2020nc",
			},
		},
	}

	if needsHDRColorMetadataFallbackFromDetailed(detailed, "/media/movie.mkv") {
		t.Fatal("expected HDR source with known color tags to skip fallback")
	}
}
