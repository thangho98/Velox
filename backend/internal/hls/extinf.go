package hls

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ParseMediaPlaylist parses ffmpeg's EVENT playlist and returns a map of segNum -> duration.
// It ignores partial/incomplete lines written due to race conditions.
func ParseMediaPlaylist(content []byte) (map[int]float64, error) {
	result := make(map[int]float64)
	scanner := bufio.NewScanner(bytes.NewReader(content))

	var currentExtinf float64
	var hasExtinf bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			parts := strings.Split(strings.TrimPrefix(line, "#EXTINF:"), ",")
			if len(parts) > 0 {
				val, err := strconv.ParseFloat(parts[0], 64)
				if err == nil {
					currentExtinf = val
					hasExtinf = true
				}
			}
			continue
		}

		if !strings.HasPrefix(line, "#") && hasExtinf {
			// Extract numerical suffix from segment filename, e.g. "v2_ss..._video_5.m4s"
			parts := strings.Split(line, "_")
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				numPart := strings.TrimSuffix(lastPart, ".m4s")
				num, err := strconv.Atoi(numPart)
				if err == nil {
					result[num] = currentExtinf
				}
			}
			hasExtinf = false
		}
	}

	return result, scanner.Err()
}

// MergeExtinf merges newly parsed segment durations into the existing map.
func MergeExtinf(existing, new map[int]float64) {
	for k, v := range new {
		existing[k] = v // overwrite with newest encode pass
	}
}

type RenderOpts struct {
	ExtinfMap       map[int]float64
	SessionStartSeg int
	TotalDuration   float64
	SegLength       float64
	Prefix          string
	Kind            string
}

// RenderMediaPlaylist generates an EVENT or VOD playlist depending on playback state.
func RenderMediaPlaylist(opts RenderOpts) string {
	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	sb.WriteString("#EXT-X-VERSION:6\n")

	targetDuration := math.Ceil(opts.SegLength)
	if targetDuration == 0 {
		targetDuration = 6
	}
	for _, dur := range opts.ExtinfMap {
		if math.Ceil(dur) > targetDuration {
			targetDuration = math.Ceil(dur)
		}
	}

	sb.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(targetDuration)))
	sb.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", opts.SessionStartSeg))
	sb.WriteString("#EXT-X-START:TIME-OFFSET=0\n") // Prevent HLS.js from auto-seeking to Live Edge
	sb.WriteString(fmt.Sprintf("#EXT-X-MAP:URI=\"%s\"\n", InitFilename(opts.Prefix, opts.Kind)))

	// Find the longest contiguous range starting from SessionStartSeg
	contiguousEnd := opts.SessionStartSeg - 1
	var keys []int
	for k := range opts.ExtinfMap {
		if k >= opts.SessionStartSeg {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)

	for _, k := range keys {
		if k == contiguousEnd+1 {
			contiguousEnd = k
		} else {
			break
		}
	}

	for i := opts.SessionStartSeg; i <= contiguousEnd; i++ {
		dur := opts.ExtinfMap[i]
		sb.WriteString(fmt.Sprintf("#EXTINF:%.6f,\n", dur))
		sb.WriteString(fmt.Sprintf("%s\n", SegmentFilename(opts.Prefix, opts.Kind, i)))
	}

	lastSegNum := int(math.Ceil(opts.TotalDuration/opts.SegLength)) - 1
	isComplete := contiguousEnd >= lastSegNum

	if isComplete {
		sb.WriteString("#EXT-X-ENDLIST\n")
	} else {
		sb.WriteString("#EXT-X-PLAYLIST-TYPE:EVENT\n")
	}

	return sb.String()
}
