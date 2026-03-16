// Package translate provides subtitle translation using DeepL (primary) and
// Google Translate (fallback). On-demand translation for SRT/VTT subtitles.
package translate

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

// Translator translates text between languages.
type Translator interface {
	// Translate translates a batch of text strings to the target language.
	// Returns translated strings in the same order.
	Translate(ctx context.Context, texts []string, targetLang string) ([]string, error)
	Name() string
}

// SRTCue represents a single subtitle cue from an SRT file.
type SRTCue struct {
	Index  string // "1", "2", etc.
	Timing string // "00:01:23,456 --> 00:01:25,789"
	Text   string // Can be multi-line
}

// ParseSRT parses an SRT file into cues.
func ParseSRT(content string) []SRTCue {
	var cues []SRTCue
	scanner := bufio.NewScanner(strings.NewReader(content))

	var current SRTCue
	state := 0 // 0=index, 1=timing, 2=text

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		switch state {
		case 0: // Expecting index
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			current.Index = trimmed
			state = 1

		case 1: // Expecting timing
			if strings.Contains(line, "-->") {
				current.Timing = line
				state = 2
			}

		case 2: // Collecting text lines
			if strings.TrimSpace(line) == "" {
				if current.Text != "" {
					cues = append(cues, current)
				}
				current = SRTCue{}
				state = 0
			} else {
				if current.Text != "" {
					current.Text += "\n"
				}
				current.Text += line
			}
		}
	}

	// Don't forget the last cue
	if current.Text != "" {
		cues = append(cues, current)
	}

	return cues
}

// BuildSRT reconstructs an SRT file from cues.
func BuildSRT(cues []SRTCue) string {
	var sb strings.Builder
	for i, cue := range cues {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(cue.Index)
		sb.WriteString("\n")
		sb.WriteString(cue.Timing)
		sb.WriteString("\n")
		sb.WriteString(cue.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// TranslateSRT translates an SRT file content to the target language.
// Batches text to minimize API calls.
func TranslateSRT(ctx context.Context, translator Translator, srtContent string, targetLang string) (string, error) {
	cues := ParseSRT(srtContent)
	if len(cues) == 0 {
		return "", fmt.Errorf("no subtitle cues found")
	}

	// Batch translate — group cues into batches of ~50 for efficiency
	const batchSize = 50
	for i := 0; i < len(cues); i += batchSize {
		end := i + batchSize
		if end > len(cues) {
			end = len(cues)
		}

		texts := make([]string, end-i)
		for j := i; j < end; j++ {
			texts[j-i] = cues[j].Text
		}

		translated, err := translator.Translate(ctx, texts, targetLang)
		if err != nil {
			return "", fmt.Errorf("translating batch %d-%d: %w", i, end, err)
		}

		for j, t := range translated {
			cues[i+j].Text = t
		}
	}

	return BuildSRT(cues), nil
}
