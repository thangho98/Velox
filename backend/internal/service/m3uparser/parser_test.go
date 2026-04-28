package m3uparser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseExtInf(t *testing.T) {
	line := `#EXTINF:-1 tvg-id="VTV1" tvg-logo="http://logo.com/vtv1.png" group-title="News" tvg-country="VN",VTV 1 HD`

	ch := parseExtInf(line)

	if ch.Name != "VTV 1 HD" {
		t.Errorf("Expected name 'VTV 1 HD', got '%s'", ch.Name)
	}
	if ch.ChannelID != "VTV1" {
		t.Errorf("Expected tvg-id 'VTV1', got '%s'", ch.ChannelID)
	}
	if ch.Logo != "http://logo.com/vtv1.png" {
		t.Errorf("Expected logo 'http://logo.com/vtv1.png', got '%s'", ch.Logo)
	}
	if ch.GroupTitle != "News" {
		t.Errorf("Expected group 'News', got '%s'", ch.GroupTitle)
	}
	if ch.Country != "VN" {
		t.Errorf("Expected country 'VN', got '%s'", ch.Country)
	}
}

func TestParser_Parse(t *testing.T) {
	m3uData := `#EXTM3U
#EXTINF:-1 tvg-id="CNN" tvg-logo="cnn.png" group-title="News",CNN International
http://stream.com/cnn.m3u8

#EXTINF:-1 group-title="Sports",ESPN
# This is a comment
http://stream.com/espn.m3u8
`

	p := New()
	channels, _, err := p.Parse(strings.NewReader(m3uData), "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(channels) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(channels))
	}

	c1 := channels[0]
	if c1.Name != "CNN International" || c1.StreamURL != "http://stream.com/cnn.m3u8" || c1.GroupTitle != "News" {
		t.Errorf("C1 parsed incorrectly: %+v", c1)
	}

	c2 := channels[1]
	if c2.Name != "ESPN" || c2.StreamURL != "http://stream.com/espn.m3u8" || c2.GroupTitle != "Sports" {
		t.Errorf("C2 parsed incorrectly: %+v", c2)
	}
}

func TestParser_ParseWithBaseURL(t *testing.T) {
	m3uData := `#EXTM3U
#EXTINF:-1 group-title="News",Relative Channel
streams/news.m3u8
`
	p := New()
	channels, _, err := p.Parse(strings.NewReader(m3uData), "https://iptv-org.github.io/iptv/countries/vn.m3u")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(channels) != 1 {
		t.Fatalf("Expected 1 channel, got %d", len(channels))
	}

	c1 := channels[0]

	// Test Country inference
	if c1.Country != "VN" {
		t.Errorf("Expected country 'VN' inferred from URL, got '%s'", c1.Country)
	}

	// Test URL normalization
	expectedURL := "https://iptv-org.github.io/iptv/countries/streams/news.m3u8"
	if c1.StreamURL != expectedURL {
		t.Errorf("Expected normalized URL '%s', got '%s'", expectedURL, c1.StreamURL)
	}
}

func TestParser_ExtractsEXTVLCOPT(t *testing.T) {
	m3uData := `#EXTM3U
#EXTINF:-1 tvg-id="mytv" group-title="VN",MyTV Channel
#EXTVLCOPT:http-user-agent="MyTV"
#EXTVLCOPT:http-referrer="https://mytv.vn/"
https://stream.example.com/mytv.m3u8
#EXTINF:-1 group-title="VN",Plain Channel
https://stream.example.com/plain.m3u8
`
	channels, _, err := New().Parse(strings.NewReader(m3uData), "")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}

	c1 := channels[0]
	if c1.UserAgent != "MyTV" {
		t.Errorf("UserAgent = %q, want MyTV", c1.UserAgent)
	}
	if c1.Referer != "https://mytv.vn/" {
		t.Errorf("Referer = %q, want https://mytv.vn/", c1.Referer)
	}

	// Second channel has no EXTVLCOPT — fields stay empty (proxy falls back to default UA).
	c2 := channels[1]
	if c2.UserAgent != "" || c2.Referer != "" || c2.Origin != "" {
		t.Errorf("expected empty headers on plain channel, got %+v", c2)
	}
}

func TestParser_ParseURL_Non200(t *testing.T) {
	// Create a dummy HTTP server that always returns 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	p := New()
	_, _, err := p.ParseURL(context.Background(), ts.URL)
	if err == nil {
		t.Fatalf("Expected error for 404 response, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected HTTP status") {
		t.Errorf("Expected error to contain 'unexpected HTTP status', got '%v'", err)
	}
}
