package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{"data": data}); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func parseID(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(r.PathValue(param), 10, 64)
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func parseInt64Query(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// resolvePublicHost returns the host:port the client used to reach the server.
// Behind a reverse proxy the inner Go server's r.Host is whatever the proxy
// forwarded as Host — nginx's default `$host` strips the port. Prefer
// X-Forwarded-Host (proxies should set it from `$http_host`) and compose with
// X-Forwarded-Port if that header lacks a port. Falls back to r.Host for
// direct connections.
func resolvePublicHost(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	// If the chosen host has no port and the proxy advertised one, append it.
	if !strings.Contains(host, ":") {
		if port := r.Header.Get("X-Forwarded-Port"); port != "" && port != "80" && port != "443" {
			host = host + ":" + port
		}
	}
	return host
}

// clientIP returns the original client IP for logging. Behind a reverse proxy
// r.RemoteAddr is the proxy IP (e.g. 127.0.0.1 for the in-container nginx).
// Order: X-Real-IP (set by trusted proxy), then leftmost X-Forwarded-For entry,
// then r.RemoteAddr (host portion of host:port). Bare IPs only — no port.
//
// Trust note: The Velox-shipped nginx in Dockerfile uses `set_real_ip_from`
// limited to private network ranges, so X-Real-IP cannot be spoofed by direct
// internet clients hitting the container. If you front Velox with another
// reverse proxy, make sure that proxy strips/overrides client-supplied
// X-Forwarded-* headers.
func clientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// Leftmost entry is the original client (subsequent are proxies).
		if comma := strings.IndexByte(v, ','); comma >= 0 {
			v = v[:comma]
		}
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
