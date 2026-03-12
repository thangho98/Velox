package playback

import (
	"testing"
)

// noMKVProfile is a Chrome-like profile that does not support MKV.
// Used for tests that need the container-incompatible branch.
var noMKVProfile = DeviceProfile{
	Name:                 "Chrome No-MKV",
	SupportedVideoCodecs: []string{CodecH264, CodecVP9, CodecAV1},
	SupportedAudioCodecs: []string{CodecAAC, CodecOpus, CodecMP3},
	SupportedContainers:  []string{ContainerMP4, ContainerWebM, ContainerHLS},
	MaxHeight:            0,
	MaxBitrate:           0,
	SupportsHLS:          true,
}

func TestDecide(t *testing.T) {
	defaultPrefs := UserPreferences{
		MaxStreamingQuality: "original",
		PreferDirectPlay:    true,
	}

	tests := []struct {
		name         string
		media        MediaFileInfo
		profile      *DeviceProfile
		prefs        UserPreferences
		wantMethod   PlaybackMethod
		wantSubtitle SubtitleAction
	}{
		{
			name: "1 H.264+AAC MP4 Chrome → DirectPlay",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "aac", Container: "mp4",
				Width: 1920, Height: 1080, Bitrate: 8000,
			},
			profile:    &ChromeDesktop,
			prefs:      defaultPrefs,
			wantMethod: MethodDirectPlay,
		},
		{
			name: "2 HEVC+AAC MP4 Chrome → FullTranscode (HEVC not supported)",
			media: MediaFileInfo{
				VideoCodec: "hevc", AudioCodec: "aac", Container: "mp4",
				Width: 1920, Height: 1080, Bitrate: 8000,
			},
			profile:    &ChromeDesktop,
			prefs:      defaultPrefs,
			wantMethod: MethodFullTranscode,
		},
		{
			name: "3 H.264+AAC MKV no-MKV profile → DirectStream (container incompatible)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "aac", Container: "mkv",
				Width: 1920, Height: 1080, Bitrate: 8000,
			},
			profile:    &noMKVProfile,
			prefs:      defaultPrefs,
			wantMethod: MethodDirectStream,
		},
		{
			name: "4 H.264+DTS MKV no-MKV profile → FullTranscode (DirectStream+audio incompatible collapses)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "dts", Container: "mkv",
				Width: 1920, Height: 1080, Bitrate: 8000,
			},
			profile:    &noMKVProfile,
			prefs:      defaultPrefs,
			wantMethod: MethodFullTranscode,
		},
		{
			name: "5 H.264+DTS MP4 Chrome → TranscodeAudio (video OK, audio incompatible)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "dts", Container: "mp4",
				Width: 1920, Height: 1080, Bitrate: 8000,
			},
			profile:    &ChromeDesktop,
			prefs:      defaultPrefs,
			wantMethod: MethodTranscodeAudio,
		},
		{
			name: "6 H.264+AAC MP4 4K with MaxHeight=1080 → FullTranscode (resolution limit)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "aac", Container: "mp4",
				Width: 3840, Height: 2160, Bitrate: 25000,
			},
			profile: &ChromeDesktop,
			prefs: UserPreferences{
				MaxStreamingQuality: "1080p",
				PreferDirectPlay:    true,
			},
			wantMethod: MethodFullTranscode,
		},
		{
			name: "7 H.264+AAC MP4 high bitrate with MaxBitrate=5000 → FullTranscode (bitrate limit)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "aac", Container: "mp4",
				Width: 1920, Height: 1080, Bitrate: 30000,
			},
			profile: &DeviceProfile{
				Name:                 "Limited",
				SupportedVideoCodecs: []string{CodecH264},
				SupportedAudioCodecs: []string{CodecAAC},
				SupportedContainers:  []string{ContainerMP4},
				MaxBitrate:           5000,
			},
			prefs:      defaultPrefs,
			wantMethod: MethodFullTranscode,
		},
		{
			name: "8 H.264+AAC MP4 PGS sub selected → FullTranscode (burn-in required)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "aac", Container: "mp4",
				Width: 1920, Height: 1080, Bitrate: 8000,
				HasSubtitles: true, SubType: SubtitlePGS,
			},
			profile: &ChromeDesktop,
			prefs: UserPreferences{
				MaxStreamingQuality: "original",
				PreferDirectPlay:    true,
				SelectedSubtitle:    "en",
			},
			wantMethod:   MethodFullTranscode,
			wantSubtitle: SubtitleBurnIn,
		},
		{
			name: "9 H.264+AAC MP4 SRT sub selected → DirectPlay + SubtitleCopy (text sub, no transcode)",
			media: MediaFileInfo{
				VideoCodec: "h264", AudioCodec: "aac", Container: "mp4",
				Width: 1920, Height: 1080, Bitrate: 8000,
				HasSubtitles: true, SubType: SubtitleSRT,
			},
			profile: &ChromeDesktop,
			prefs: UserPreferences{
				MaxStreamingQuality: "original",
				PreferDirectPlay:    true,
				SelectedSubtitle:    "en",
			},
			wantMethod:   MethodDirectPlay,
			wantSubtitle: SubtitleCopy,
		},
		{
			name: "10 HEVC+AC3 MKV Safari → DirectPlay (Safari supports HEVC+AC3+MKV)",
			media: MediaFileInfo{
				VideoCodec: "hevc", AudioCodec: "ac-3", Container: "mkv",
				Width: 1920, Height: 1080, Bitrate: 8000,
			},
			profile:    &SafariDesktop,
			prefs:      defaultPrefs,
			wantMethod: MethodDirectPlay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := Decide(tt.media, tt.profile, tt.prefs)

			if decision.Method != tt.wantMethod {
				t.Errorf("Method = %q, want %q (reason: %s)", decision.Method, tt.wantMethod, decision.Reason)
			}

			if tt.wantSubtitle != "" && decision.SubtitleAction != tt.wantSubtitle {
				t.Errorf("SubtitleAction = %q, want %q", decision.SubtitleAction, tt.wantSubtitle)
			}
		})
	}
}
