package podnapisi

import (
	"encoding/xml"

	"github.com/thawng/velox/pkg/subprovider"
)

// XML response types for Podnapisi sXML=1 endpoint

type xmlResponse struct {
	XMLName   xml.Name      `xml:"results"`
	Subtitles []xmlSubtitle `xml:"subtitle"`
}

type xmlSubtitle struct {
	PID        string  `xml:"pid"`
	Title      string  `xml:"title"`
	Release    string  `xml:"release"`
	LanguageID string  `xml:"languageId"`
	Language   string  `xml:"languageName"`
	Downloads  int     `xml:"downloads"`
	Rating     float64 `xml:"rating"`
	Flags      string  `xml:"flags"`
}

func parseXMLResponse(data []byte) ([]subprovider.Result, error) {
	var resp xmlResponse
	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	results := make([]subprovider.Result, 0, len(resp.Subtitles))
	for _, sub := range resp.Subtitles {
		title := sub.Release
		if title == "" {
			title = sub.Title
		}

		results = append(results, subprovider.Result{
			Provider:        "podnapisi",
			ExternalID:      sub.PID,
			Title:           title,
			Language:        reverseLangCode(sub.LanguageID),
			Format:          "srt",
			Downloads:       sub.Downloads,
			Rating:          sub.Rating,
			HearingImpaired: containsFlag(sub.Flags, "n"),
		})
	}
	return results, nil
}

func containsFlag(flags, flag string) bool {
	for _, f := range flags {
		if string(f) == flag {
			return true
		}
	}
	return false
}
