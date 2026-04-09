package hls

import (
	"fmt"
	"strings"

	"github.com/thawng/velox/internal/model"
)

func GenerateMasterPlaylist(audioTracks []model.AudioTrack, prefix string, videoBandwidth int) string {
	var sb strings.Builder

	sb.WriteString("#EXTM3U\n")
	sb.WriteString("#EXT-X-VERSION:6\n")
	sb.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n\n")

	audioGroup := "audio"
	hasAudio := len(audioTracks) > 0

	for i, track := range audioTracks {
		isDefault := "NO"
		if track.IsDefault {
			isDefault = "YES"
		}
		// In v2, audio kind is audio0, audio1, etc.
		uri := fmt.Sprintf("%saudio%d.m3u8", prefix, i)
		lang := track.Language
		if lang == "" {
			lang = "und"
		}
		name := track.Title
		if name == "" {
			name = fmt.Sprintf("Track %d", i+1)
		}

		sb.WriteString(fmt.Sprintf("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"%s\",NAME=\"%s\",LANGUAGE=\"%s\",DEFAULT=%s,AUTOSELECT=YES,URI=\"%s\"\n",
			audioGroup, name, lang, isDefault, uri))
	}

	if hasAudio {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,CODECS=\"avc1.640028,mp4a.40.2\",AUDIO=\"%s\"\n", videoBandwidth, audioGroup))
	} else {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,CODECS=\"avc1.640028\"\n", videoBandwidth))
	}
	sb.WriteString(MediaPlaylistFilename(prefix, "video") + "\n")

	return sb.String()
}
