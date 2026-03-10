package ffprobe

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type ProbeResult struct {
	Duration   float64
	Width      int
	Height     int
	VideoCodec string
	AudioCodec string
	Container  string
	Bitrate    int64
	HasSub     bool
}

type probeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
		Name     string `json:"format_name"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

// Probe runs ffprobe on the given file and returns parsed metadata.
func Probe(path string) (*ProbeResult, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w", err)
	}

	var p probeOutput
	if err := json.Unmarshal(out, &p); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	r := &ProbeResult{
		Container: p.Format.Name,
	}

	r.Duration, _ = strconv.ParseFloat(p.Format.Duration, 64)
	r.Bitrate, _ = strconv.ParseInt(p.Format.BitRate, 10, 64)

	for _, s := range p.Streams {
		switch s.CodecType {
		case "video":
			if r.VideoCodec == "" {
				r.VideoCodec = s.CodecName
				r.Width = s.Width
				r.Height = s.Height
			}
		case "audio":
			if r.AudioCodec == "" {
				r.AudioCodec = s.CodecName
			}
		case "subtitle":
			r.HasSub = true
		}
	}

	return r, nil
}
