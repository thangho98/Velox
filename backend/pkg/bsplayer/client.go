package bsplayer

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thawng/velox/pkg/subprovider"
)

const (
	bsUserAgent = "BSPlayer/2.x (1022.12362)"
	bsAppID     = "BSPlayer v2.72"
)

var subdomains = []int{1, 2, 3, 4, 5, 6, 7, 8, 101, 102, 103, 104, 105, 106, 107, 108, 109}

// Client is a BSPlayer subtitle API client (SOAP/XML). No API key required.
type Client struct {
	http      *http.Client
	subdomain int
}

// SearchParams configures a BSPlayer subtitle search query.
type SearchParams struct {
	FileHash string // BSPlayer/OpenSubtitles-style hash (optional, "0" if unknown)
	FileSize int64  // file size in bytes (optional, 0 if unknown)
	ImdbID   string // IMDb ID (e.g. "tt1234567" or "1234567")
	Language string // ISO 639-1 (e.g. "en")
}

// New creates a new BSPlayer client with a random subdomain.
func New() *Client {
	sd := subdomains[rand.IntN(len(subdomains))]
	return &Client{
		http:      &http.Client{Timeout: 30 * time.Second},
		subdomain: sd,
	}
}

func (c *Client) apiURL() string {
	return fmt.Sprintf("http://s%d.api.bsplayer-subtitles.com/v1.php", c.subdomain)
}

// Search performs login, subtitle search, and logout in one call.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	token, err := c.login(ctx)
	if err != nil {
		return nil, fmt.Errorf("bsplayer login: %w", err)
	}
	defer c.logout(token)

	langCode := isoToBSLang(params.Language)

	hash := params.FileHash
	if hash == "" {
		hash = "0"
	}
	size := "0"
	if params.FileSize > 0 {
		size = strconv.FormatInt(params.FileSize, 10)
	}

	imdbID := params.ImdbID
	if strings.HasPrefix(imdbID, "tt") {
		imdbID = imdbID[2:]
	}
	if imdbID == "" {
		imdbID = "0"
	}

	searchXML := fmt.Sprintf(
		"<handle>%s</handle>"+
			"<movieHash>%s</movieHash>"+
			"<movieSize>%s</movieSize>"+
			"<languageId>%s</languageId>"+
			"<imdbId>%s</imdbId>",
		token, hash, size, langCode, imdbID,
	)

	body, err := c.soapRequest(ctx, "searchSubtitles", searchXML)
	if err != nil {
		return nil, fmt.Errorf("bsplayer search: %w", err)
	}

	return parseSearchResponse(body, params.Language)
}

// Download fetches a subtitle file. BSPlayer typically returns gzip-compressed data.
func (c *Client) Download(ctx context.Context, downloadURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", bsUserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("bsplayer download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("bsplayer download failed (%d): %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading download body: %w", err)
	}

	// Try gzip decompress (BSPlayer responses are often gzipped)
	gr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		// Not gzipped — return as-is
		return body, "subtitle.srt", nil
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, "", fmt.Errorf("decompressing subtitle: %w", err)
	}

	return data, "subtitle.srt", nil
}

func (c *Client) login(ctx context.Context) (string, error) {
	params := "<username></username><password></password><AppID>" + bsAppID + "</AppID>"
	body, err := c.soapRequest(ctx, "logIn", params)
	if err != nil {
		return "", err
	}

	ret, err := extractReturnElement(body)
	if err != nil {
		return "", fmt.Errorf("parsing login response: %w", err)
	}

	var lr loginReturn
	if err := xml.Unmarshal(ret, &lr); err != nil {
		return "", fmt.Errorf("decoding login return: %w", err)
	}

	if lr.Result != "200" {
		return "", fmt.Errorf("login failed: status %s", lr.Result)
	}
	if lr.Data == "" {
		return "", fmt.Errorf("login returned empty token")
	}
	return lr.Data, nil
}

func (c *Client) logout(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	params := "<handle>" + token + "</handle>"
	_, _ = c.soapRequest(ctx, "logOut", params)
}

func (c *Client) soapRequest(ctx context.Context, action, params string) ([]byte, error) {
	apiURL := c.apiURL()
	soapBody := fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>`+
			`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" `+
			`xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" `+
			`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" `+
			`xmlns:xsd="http://www.w3.org/2001/XMLSchema" `+
			`xmlns:ns1="%s">`+
			`<SOAP-ENV:Body SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`+
			`<ns1:%s>%s</ns1:%s>`+
			`</SOAP-ENV:Body>`+
			`</SOAP-ENV:Envelope>`,
		apiURL, action, params, action,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(soapBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", bsUserAgent)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", fmt.Sprintf(`"%s#%s"`, apiURL, action))
	req.Header.Set("Connection", "close")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	return io.ReadAll(resp.Body)
}

// extractReturnElement finds the <return>...</return> element in SOAP response.
func extractReturnElement(data []byte) ([]byte, error) {
	s := string(data)
	start := strings.Index(s, "<return")
	if start == -1 {
		return nil, fmt.Errorf("no <return> element in response")
	}
	end := strings.LastIndex(s, "</return>")
	if end == -1 {
		return nil, fmt.Errorf("no </return> closing tag in response")
	}
	return []byte(s[start : end+len("</return>")]), nil
}

func parseSearchResponse(data []byte, isoLang string) ([]subprovider.Result, error) {
	ret, err := extractReturnElement(data)
	if err != nil {
		return nil, err
	}

	// Extract status code from <result>
	statusCode := extractStatusCode(ret)
	if statusCode != "200" && statusCode != "402" {
		return nil, fmt.Errorf("search failed: status %q", statusCode)
	}

	var sr searchReturn
	if err := xml.Unmarshal(ret, &sr); err != nil {
		return nil, fmt.Errorf("decoding search return: %w", err)
	}

	results := make([]subprovider.Result, 0, len(sr.Data.Items))
	for _, item := range sr.Data.Items {
		if item.SubDownloadLink == "" {
			continue
		}

		var rating float64
		if item.SubRating != "" {
			r, _ := strconv.ParseFloat(item.SubRating, 64)
			rating = r / 2 // BSPlayer rates 0-10, normalize to 0-5
		}

		lang := bsLangToISO(item.SubLang)
		if lang == "" {
			lang = isoLang
		}

		results = append(results, subprovider.Result{
			Provider:   "bsplayer",
			ExternalID: item.SubDownloadLink, // download URL is the ID
			Title:      item.SubName,
			Language:   lang,
			Format:     "srt",
			Rating:     rating,
		})
	}
	return results, nil
}

// XML types for SOAP response parsing

type loginReturn struct {
	XMLName xml.Name `xml:"return"`
	Result  string   `xml:"result"`
	Data    string   `xml:"data"`
}

type searchReturn struct {
	XMLName xml.Name `xml:"return"`
	Data    struct {
		Items []searchItem `xml:"item"`
	} `xml:"data"`
}

// extractStatusCode pulls the status code from the SOAP return XML.
// BSPlayer uses xsi:type attributes, so tags look like:
//
//	<result xsi:type="ns1:SubtitlesResult"><result xsi:type="xsd:string">402</result>...
//
// We find the innermost <result ...>CODE</result> where CODE is a short numeric string.
func extractStatusCode(data []byte) string {
	s := string(data)
	tag := "result"
	searchFrom := 0
	for {
		// Find next <result with possible attributes
		idx := strings.Index(s[searchFrom:], "<"+tag)
		if idx == -1 {
			break
		}
		idx += searchFrom
		// Find the closing >
		gtIdx := strings.Index(s[idx:], ">")
		if gtIdx == -1 {
			break
		}
		contentStart := idx + gtIdx + 1
		// Find </result>
		closeIdx := strings.Index(s[contentStart:], "</"+tag+">")
		if closeIdx == -1 {
			break
		}
		content := s[contentStart : contentStart+closeIdx]
		// If content is short and doesn't contain child elements, it's the status code
		if len(content) <= 5 && !strings.Contains(content, "<") {
			return strings.TrimSpace(content)
		}
		searchFrom = contentStart
	}
	return ""
}

type searchItem struct {
	SubName         string `xml:"subName"`
	SubLang         string `xml:"subLang"`
	SubRating       string `xml:"subRating"`
	SubDownloadLink string `xml:"subDownloadLink"`
}

// Language mapping: ISO 639-1 → ISO 639-2/B (BSPlayer uses 3-letter codes)
var langISO1ToBS = map[string]string{
	"en": "eng", "vi": "vie", "fr": "fre", "de": "ger",
	"es": "spa", "pt": "por", "it": "ita", "nl": "dut",
	"pl": "pol", "ru": "rus", "ja": "jpn", "ko": "kor",
	"zh": "chi", "ar": "ara", "tr": "tur", "sv": "swe",
	"da": "dan", "fi": "fin", "no": "nor", "cs": "cze",
	"hu": "hun", "ro": "rum", "hr": "hrv", "sr": "srp",
	"bg": "bul", "el": "gre", "he": "heb", "th": "tha",
	"id": "ind", "ms": "may",
}

func isoToBSLang(iso string) string {
	if code, ok := langISO1ToBS[strings.ToLower(iso)]; ok {
		return code
	}
	return iso
}

func bsLangToISO(bsLang string) string {
	lower := strings.ToLower(bsLang)
	for iso, bs := range langISO1ToBS {
		if bs == lower {
			return iso
		}
	}
	return bsLang
}
