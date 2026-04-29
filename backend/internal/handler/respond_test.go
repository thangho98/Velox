package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondJSON(t *testing.T) {
	t.Parallel()

	t.Run("writes data with data wrapper", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		respondJSON(w, http.StatusOK, map[string]string{"hello": "world"})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		data, ok := resp["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "world", data["hello"])
	})

	t.Run("writes struct", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		type resp struct{ Name string }
		respondJSON(w, http.StatusOK, resp{Name: "test"})

		assert.Equal(t, http.StatusOK, w.Code)
		var respMap map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &respMap)
		require.NoError(t, err)
	})

	t.Run("StatusNoContent omits body", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		respondJSON(w, http.StatusNoContent, nil)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.Len())
	})
}

func TestRespondError(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	respondError(w, http.StatusBadRequest, "invalid id")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid id", resp["error"])
}

func TestParseID(t *testing.T) {
	t.Parallel()

	t.Run("valid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", "123")
		id, err := parseID(req, "id")
		assert.NoError(t, err)
		assert.Equal(t, int64(123), id)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", "abc")
		_, err := parseID(req, "id")
		assert.Error(t, err)
	})
}

func TestParseIntQuery(t *testing.T) {
	t.Parallel()

	t.Run("default when missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		got := parseIntQuery(req, "limit", 50)
		assert.Equal(t, 50, got)
	})

	t.Run("parses value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?limit=25", nil)
		got := parseIntQuery(req, "limit", 50)
		assert.Equal(t, 25, got)
	})

	t.Run("fallback on invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?limit=abc", nil)
		got := parseIntQuery(req, "limit", 50)
		assert.Equal(t, 50, got)
	})
}

func TestParseInt64Query(t *testing.T) {
	t.Parallel()

	t.Run("valid string", func(t *testing.T) {
		v, err := parseInt64Query("12345")
		assert.NoError(t, err)
		assert.Equal(t, int64(12345), v)
	})

	t.Run("invalid string", func(t *testing.T) {
		_, err := parseInt64Query("abc")
		assert.Error(t, err)
	})
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	t.Run("existing file", func(t *testing.T) {
		exists := fileExists("/etc/passwd")
		assert.True(t, exists)
	})

	t.Run("non-existing file", func(t *testing.T) {
		exists := fileExists("/nonexistent/path/to/file")
		assert.False(t, exists)
	})
}

func TestClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		remoteAddr    string
		xRealIP       string
		xForwardedFor string
		want          string
	}{
		{
			name:       "direct connection strips port from RemoteAddr",
			remoteAddr: "1.2.3.4:54321",
			want:       "1.2.3.4",
		},
		{
			name:       "RemoteAddr without port is returned as-is",
			remoteAddr: "1.2.3.4",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Real-IP wins over RemoteAddr",
			remoteAddr: "127.0.0.1:54321",
			xRealIP:    "1.2.3.4",
			want:       "1.2.3.4",
		},
		{
			name:          "X-Forwarded-For leftmost wins when X-Real-IP missing",
			remoteAddr:    "127.0.0.1:54321",
			xForwardedFor: "1.2.3.4, 10.0.0.1, 172.17.0.2",
			want:          "1.2.3.4",
		},
		{
			name:          "X-Real-IP takes priority over X-Forwarded-For",
			remoteAddr:    "127.0.0.1:54321",
			xRealIP:       "1.2.3.4",
			xForwardedFor: "9.9.9.9, 10.0.0.1",
			want:          "1.2.3.4",
		},
		{
			name:       "IPv6 RemoteAddr strips port and brackets",
			remoteAddr: "[2001:db8::1]:54321",
			want:       "2001:db8::1",
		},
		{
			name:          "trims whitespace from X-Forwarded-For entries",
			remoteAddr:    "127.0.0.1:54321",
			xForwardedFor: "   1.2.3.4   , 10.0.0.1",
			want:          "1.2.3.4",
		},
		{
			name:       "trims whitespace from X-Real-IP",
			remoteAddr: "127.0.0.1:54321",
			xRealIP:    "  1.2.3.4  ",
			want:       "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			got := clientIP(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolvePublicHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		reqHost        string
		xForwardedHost string
		xForwardedPort string
		want           string
	}{
		{
			name:    "direct connection preserves host with port",
			reqHost: "192.168.98.98:8098",
			want:    "192.168.98.98:8098",
		},
		{
			name:    "direct connection with bare host",
			reqHost: "velox.local",
			want:    "velox.local",
		},
		{
			name:           "X-Forwarded-Host with port wins over r.Host",
			reqHost:        "127.0.0.1:8080",
			xForwardedHost: "192.168.98.98:8098",
			want:           "192.168.98.98:8098",
		},
		{
			name:           "X-Forwarded-Host without port composes with X-Forwarded-Port",
			reqHost:        "127.0.0.1:8080",
			xForwardedHost: "velox.local",
			xForwardedPort: "8098",
			want:           "velox.local:8098",
		},
		{
			name:           "r.Host without port composes with X-Forwarded-Port",
			reqHost:        "192.168.98.98",
			xForwardedPort: "8098",
			want:           "192.168.98.98:8098",
		},
		{
			name:           "X-Forwarded-Port=80 is omitted (default http)",
			reqHost:        "velox.local",
			xForwardedPort: "80",
			want:           "velox.local",
		},
		{
			name:           "X-Forwarded-Port=443 is omitted (default https)",
			reqHost:        "velox.local",
			xForwardedPort: "443",
			want:           "velox.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.reqHost
			if tt.xForwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tt.xForwardedHost)
			}
			if tt.xForwardedPort != "" {
				req.Header.Set("X-Forwarded-Port", tt.xForwardedPort)
			}
			got := resolvePublicHost(req)
			assert.Equal(t, tt.want, got)
		})
	}
}
