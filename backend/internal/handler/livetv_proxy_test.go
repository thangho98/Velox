package handler

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newTestLiveTVHandler(secret []byte) *LiveTVHandler {
	return &LiveTVHandler{
		proxySecret: secret,
		proxyClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestRewriteManifest_MasterPlaylist(t *testing.T) {
	h := newTestLiveTVHandler([]byte("test-secret-32-bytes-long-xxxxxx"))

	manifest := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=640x360
720p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=3840000,RESOLUTION=1280x720
https://cdn.example.com/1080p/playlist.m3u8
`
	base := "https://cdn.example.com/channels/abc/master.m3u8"
	out, err := h.rewriteManifest(manifest, base, upstreamHeaders{})
	if err != nil {
		t.Fatalf("rewriteManifest: %v", err)
	}

	// Every non-tag line should be rewritten to /api/livetv/proxy?...
	if strings.Contains(out, "720p/playlist.m3u8\n") {
		t.Errorf("relative child URL not rewritten:\n%s", out)
	}
	if strings.Contains(out, "https://cdn.example.com/1080p/playlist.m3u8\n") {
		t.Errorf("absolute child URL not rewritten:\n%s", out)
	}
	if !strings.Contains(out, "/api/livetv/proxy?") {
		t.Errorf("expected proxy URLs in output:\n%s", out)
	}

	// EXT-X-STREAM-INF lines (tags) should be preserved verbatim.
	if !strings.Contains(out, "#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=640x360") {
		t.Errorf("stream-inf tag missing:\n%s", out)
	}
}

func TestRewriteManifest_URIAttribute(t *testing.T) {
	h := newTestLiveTVHandler([]byte("test-secret"))

	manifest := `#EXTM3U
#EXT-X-KEY:METHOD=AES-128,URI="key.bin",IV=0x00
#EXT-X-MAP:URI="init.mp4"
seg1.ts
`
	base := "https://cdn.example.com/channels/abc/playlist.m3u8"
	out, err := h.rewriteManifest(manifest, base, upstreamHeaders{})
	if err != nil {
		t.Fatalf("rewriteManifest: %v", err)
	}

	if strings.Contains(out, `URI="key.bin"`) {
		t.Errorf("EXT-X-KEY URI not rewritten:\n%s", out)
	}
	if strings.Contains(out, `URI="init.mp4"`) {
		t.Errorf("EXT-X-MAP URI not rewritten:\n%s", out)
	}
	// Should contain two URI= attributes pointing at the proxy
	if strings.Count(out, `URI="/api/livetv/proxy?`) != 2 {
		t.Errorf("expected 2 rewritten URI attributes:\n%s", out)
	}
}

func TestRewriteManifest_PreservesOrder(t *testing.T) {
	h := newTestLiveTVHandler([]byte("s"))

	manifest := `#EXTM3U
#EXT-X-VERSION:3
#EXTINF:10.0,
a.ts
#EXTINF:10.0,
b.ts
#EXT-X-ENDLIST
`
	base := "https://x.com/p.m3u8"
	out, err := h.rewriteManifest(manifest, base, upstreamHeaders{})
	if err != nil {
		t.Fatalf("rewriteManifest: %v", err)
	}

	lines := strings.Split(out, "\n")
	// Verify: #EXTM3U, #VERSION, #EXTINF, <proxy for a>, #EXTINF, <proxy for b>, #ENDLIST
	if lines[0] != "#EXTM3U" || lines[1] != "#EXT-X-VERSION:3" {
		t.Errorf("header malformed:\n%s", out)
	}
	if !strings.HasPrefix(lines[3], "/api/livetv/proxy?") {
		t.Errorf("first segment not rewritten at line 3: %q\n%s", lines[3], out)
	}
	if !strings.HasPrefix(lines[5], "/api/livetv/proxy?") {
		t.Errorf("second segment not rewritten at line 5: %q\n%s", lines[5], out)
	}
	if lines[6] != "#EXT-X-ENDLIST" {
		t.Errorf("ENDLIST tag not preserved:\n%s", out)
	}
}

func TestProxySignature_RoundTrip(t *testing.T) {
	secret := []byte("secret-key")
	upstream := "https://cdn.example.com/a/b.ts?token=xyz"
	exp := time.Now().Add(time.Hour).Unix()
	hdrs := upstreamHeaders{}

	sig := signProxyURL(secret, upstream, exp, hdrs)
	if !verifyProxySignature(secret, upstream, exp, hdrs, sig) {
		t.Errorf("valid signature rejected")
	}

	// Tamper URL
	if verifyProxySignature(secret, upstream+"x", exp, hdrs, sig) {
		t.Errorf("tampered URL accepted")
	}
	// Tamper exp
	if verifyProxySignature(secret, upstream, exp+1, hdrs, sig) {
		t.Errorf("tampered exp accepted")
	}
	// Wrong secret
	if verifyProxySignature([]byte("other"), upstream, exp, hdrs, sig) {
		t.Errorf("wrong secret accepted")
	}
	// Tamper headers (attacker injects a UA the server didn't sign)
	if verifyProxySignature(secret, upstream, exp, upstreamHeaders{UserAgent: "evil"}, sig) {
		t.Errorf("tampered headers accepted")
	}
}

func TestProxySignature_WithHeaders(t *testing.T) {
	secret := []byte("secret-key")
	upstream := "https://cdn.example.com/a/b.ts"
	exp := time.Now().Add(time.Hour).Unix()
	hdrs := upstreamHeaders{UserAgent: "MyTV", Referer: "https://mytv.vn/"}

	sig := signProxyURL(secret, upstream, exp, hdrs)
	if !verifyProxySignature(secret, upstream, exp, hdrs, sig) {
		t.Errorf("valid signature with headers rejected")
	}
	// Dropping a header must invalidate the signature.
	if verifyProxySignature(secret, upstream, exp, upstreamHeaders{UserAgent: "MyTV"}, sig) {
		t.Errorf("signature accepted after referer stripped")
	}
}

func TestMakeProxyURL_IsParsable(t *testing.T) {
	h := newTestLiveTVHandler([]byte("k"))
	out := h.makeProxyURL("https://cdn.example.com/x.ts?token=abc&a=b", upstreamHeaders{})

	u, err := url.Parse(out)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Path != "/api/livetv/proxy" {
		t.Errorf("wrong path: %q", u.Path)
	}
	if u.Query().Get("u") != "https://cdn.example.com/x.ts?token=abc&a=b" {
		t.Errorf("upstream url not preserved: %q", u.Query().Get("u"))
	}
	if u.Query().Get("sig") == "" || u.Query().Get("exp") == "" {
		t.Errorf("missing sig/exp: %q", out)
	}
}

func TestMakeProxyURL_EncodesHeaders(t *testing.T) {
	h := newTestLiveTVHandler([]byte("k"))
	hdrs := upstreamHeaders{UserAgent: "MyTV", Referer: "https://mytv.vn/"}
	out := h.makeProxyURL("https://cdn.example.com/x.ts", hdrs)

	u, err := url.Parse(out)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Query().Get("ua") != "MyTV" {
		t.Errorf("ua not encoded: %q", out)
	}
	if u.Query().Get("ref") != "https://mytv.vn/" {
		t.Errorf("ref not encoded: %q", out)
	}
	// Origin not set → should be omitted.
	if _, ok := u.Query()["org"]; ok {
		t.Errorf("empty origin should not be emitted: %q", out)
	}
}
