package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/repository"
)

const (
	proxySignatureTTL = 2 * time.Hour
	// defaultProxyUserAgent is used when the channel doesn't specify its own.
	// Match the most common case where upstream blocks unknown UAs.
	defaultProxyUserAgent = "VLC/3.0.18 LibVLC/3.0.18"
)

var manifestTagURIRegex = regexp.MustCompile(`URI="([^"]+)"`)

// upstreamHeaders carries the per-channel HTTP hints (#EXTVLCOPT) that origin
// CDNs require. Empty strings fall back to defaults — UserAgent falls back to
// defaultProxyUserAgent, Referer/Origin are simply omitted when empty.
type upstreamHeaders struct {
	UserAgent string
	Referer   string
	Origin    string
}

// handleStreamEntry is the public entry point that browsers load in the <video>
// element: it proxies the channel's manifest through our backend so that CORS
// and hotlink restrictions from the origin CDN are bypassed.
func (h *LiveTVHandler) handleStreamEntry(w http.ResponseWriter, r *http.Request) {
	channelIDStr := r.PathValue("channelId")
	id, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	ch, err := h.repo.GetChannelByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ch.StreamURL == "" {
		respondError(w, http.StatusBadGateway, "channel has no stream url")
		return
	}

	// Record "recently watched" when we can identify the user (JWT in header or
	// ?token= query). Skip-paths pre-parse the token; see middleware.withOptionalJWTContext.
	// Unknown users (external players) stream normally without being tracked.
	if userID, _, ok := auth.UserFromContext(r.Context()); ok && h.shouldMarkWatched(userID, ch.ID) {
		go func(uid, cid int64) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.repo.MarkChannelWatched(ctx, uid, cid); err != nil {
				log.Printf("livetv: mark channel watched user=%d channel=%d: %v", uid, cid, err)
			}
		}(userID, ch.ID)
	}

	hdrs := upstreamHeaders{UserAgent: ch.UserAgent, Referer: ch.Referer, Origin: ch.Origin}
	h.proxyFetch(w, r, ch.StreamURL, hdrs)
}

// handleProxy proxies any upstream URL that carries a valid HMAC signature.
// Signatures are minted server-side when rewriting a manifest, so clients can
// only follow links we emitted — they cannot craft arbitrary proxy targets
// (prevents SSRF). Per-channel UA/Referer/Origin hints are signed along with
// the URL so segments can carry forward the channel's required headers.
func (h *LiveTVHandler) handleProxy(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rawURL := q.Get("u")
	sig := q.Get("sig")
	expStr := q.Get("exp")
	if rawURL == "" || sig == "" || expStr == "" {
		respondError(w, http.StatusBadRequest, "missing proxy params")
		return
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid exp")
		return
	}
	if time.Now().Unix() > exp {
		respondError(w, http.StatusForbidden, "signature expired")
		return
	}
	hdrs := upstreamHeaders{
		UserAgent: q.Get("ua"),
		Referer:   q.Get("ref"),
		Origin:    q.Get("org"),
	}
	if !verifyProxySignature(h.proxySecret, rawURL, exp, hdrs, sig) {
		respondError(w, http.StatusForbidden, "invalid signature")
		return
	}
	h.proxyFetch(w, r, rawURL, hdrs)
}

// proxyFetch performs the actual upstream request and streams the response
// back. HLS manifests are rewritten so every child URL (variants + segments +
// keys + init sections) routes through this proxy.
func (h *LiveTVHandler) proxyFetch(w http.ResponseWriter, r *http.Request, upstreamURL string, hdrs upstreamHeaders) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, nil)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid upstream url")
		return
	}
	// Forward Range for byte-range segment fetches (some segments use fMP4).
	if rangeHdr := r.Header.Get("Range"); rangeHdr != "" {
		req.Header.Set("Range", rangeHdr)
	}
	if hdrs.UserAgent != "" {
		req.Header.Set("User-Agent", hdrs.UserAgent)
	} else {
		req.Header.Set("User-Agent", defaultProxyUserAgent)
	}
	if hdrs.Referer != "" {
		req.Header.Set("Referer", hdrs.Referer)
	}
	if hdrs.Origin != "" {
		req.Header.Set("Origin", hdrs.Origin)
	}
	req.Header.Set("Accept", "*/*")

	resp, err := h.proxyClient.Do(req)
	if err != nil {
		respondError(w, http.StatusBadGateway, "upstream fetch failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// CORS: let the browser read range/length for seeking.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Range")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range,Accept-Ranges")

	if isManifestResponse(resp, upstreamURL) && resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			respondError(w, http.StatusBadGateway, "read manifest failed: "+err.Error())
			return
		}
		rewritten, err := h.rewriteManifest(string(body), upstreamURL, hdrs)
		if err != nil {
			respondError(w, http.StatusBadGateway, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(rewritten))
		return
	}

	// Passthrough — forward selected headers, status, and body bytes.
	for _, key := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Cache-Control"} {
		if v := resp.Header.Get(key); v != "" {
			w.Header().Set(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// rewriteManifest replaces every URL line (and URI="..." attributes inside
// EXT-X-KEY / EXT-X-MAP / EXT-X-MEDIA tags) with a signed proxy URL resolved
// against the playlist's own base URL. The channel's upstream headers are
// carried in each signed URL so segment fetches keep the same UA/Referer.
func (h *LiveTVHandler) rewriteManifest(body, baseURL string, hdrs upstreamHeaders) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse manifest base url: %w", err)
	}

	var out strings.Builder
	out.Grow(len(body) + len(body)/4)

	// Preserve original line endings by splitting on "\n" and letting the
	// trailing "\r" (if any) pass through untouched on non-URL lines.
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "" {
			out.WriteString(line)
			if i < len(lines)-1 {
				out.WriteString("\n")
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			// Tag line — rewrite any URI="..." attribute.
			rewritten := manifestTagURIRegex.ReplaceAllStringFunc(line, func(match string) string {
				sub := manifestTagURIRegex.FindStringSubmatch(match)
				if len(sub) < 2 {
					return match
				}
				resolved, err := resolveManifestURL(base, sub[1])
				if err != nil {
					return match
				}
				return fmt.Sprintf(`URI="%s"`, h.makeProxyURL(resolved, hdrs))
			})
			out.WriteString(rewritten)
		} else {
			// Non-tag, non-blank → segment or variant URL.
			resolved, err := resolveManifestURL(base, trimmed)
			if err != nil {
				out.WriteString(line)
			} else {
				out.WriteString(h.makeProxyURL(resolved, hdrs))
			}
		}
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}
	return out.String(), nil
}

func resolveManifestURL(base *url.URL, ref string) (string, error) {
	refURL, err := url.Parse(strings.TrimSpace(ref))
	if err != nil {
		return "", err
	}
	return base.ResolveReference(refURL).String(), nil
}

func (h *LiveTVHandler) makeProxyURL(upstream string, hdrs upstreamHeaders) string {
	exp := time.Now().Add(proxySignatureTTL).Unix()
	sig := signProxyURL(h.proxySecret, upstream, exp, hdrs)
	v := url.Values{}
	v.Set("u", upstream)
	v.Set("exp", strconv.FormatInt(exp, 10))
	v.Set("sig", sig)
	// Only emit header params when non-default — keeps URLs compact for
	// channels that don't need custom upstream headers.
	if hdrs.UserAgent != "" {
		v.Set("ua", hdrs.UserAgent)
	}
	if hdrs.Referer != "" {
		v.Set("ref", hdrs.Referer)
	}
	if hdrs.Origin != "" {
		v.Set("org", hdrs.Origin)
	}
	return "/api/livetv/proxy?" + v.Encode()
}

func signProxyURL(secret []byte, upstreamURL string, exp int64, hdrs upstreamHeaders) string {
	mac := hmac.New(sha256.New, secret)
	fmt.Fprintf(mac, "%s|%d|%s|%s|%s", upstreamURL, exp, hdrs.UserAgent, hdrs.Referer, hdrs.Origin)
	return hex.EncodeToString(mac.Sum(nil))
}

func verifyProxySignature(secret []byte, upstreamURL string, exp int64, hdrs upstreamHeaders, sig string) bool {
	expected := signProxyURL(secret, upstreamURL, exp, hdrs)
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	expectedBytes, _ := hex.DecodeString(expected)
	return hmac.Equal(sigBytes, expectedBytes)
}

func isManifestResponse(resp *http.Response, upstreamURL string) bool {
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(ct, "mpegurl") || strings.Contains(ct, "m3u") {
		return true
	}
	// Some CDNs serve m3u8 as octet-stream; fall back to the path suffix.
	if u, err := url.Parse(upstreamURL); err == nil {
		if strings.HasSuffix(strings.ToLower(u.Path), ".m3u8") {
			return true
		}
	}
	return false
}
