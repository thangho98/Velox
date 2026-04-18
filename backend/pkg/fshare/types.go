package fshare

import (
	"fmt"
	"strconv"
	"time"
)

// parseInt64 parses a decimal string. Empty string returns 0 with no error.
func parseInt64(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse int64 %q: %w", s, err)
	}
	return n, nil
}

// LoginRequest is the body sent to POST /api/user/login.
type LoginRequest struct {
	UserEmail string `json:"user_email"`
	Password  string `json:"password"`
	AppKey    string `json:"app_key"`
}

// LoginResponse is the body returned from POST /api/user/login.
// Code == 200 indicates success; other values are logical errors (HTTP is still 200).
// Only {code, msg, token, session_id} are returned — email and VIP status must
// be fetched separately via /api/user/get (see UserInfo + GetUserInfo).
type LoginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
}

// UserInfo is the body returned from GET /api/user/get. Populated with account
// details including VIP status. Fields are strings in the wire format.
type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccountType string `json:"account_type"` // "Vip" | "Fee"
	ExpireVIP   string `json:"expire_vip"`   // unix timestamp, seconds
	Webspace    string `json:"webspace"`
	Level       string `json:"level"`
}

// FolderListRequest is the body sent to POST /api/fileops/getFolderList (nested folders).
type FolderListRequest struct {
	Token     string `json:"token"`
	URL       string `json:"url"` // "https://www.fshare.vn/folder/<linkcode>"
	DirOnly   int    `json:"dirOnly"`
	PageIndex int    `json:"pageIndex"`
	Limit     int    `json:"limit"`
}

// FolderItem represents a file or subfolder returned by list endpoints.
// fshare returns all fields as strings (including numeric-looking ones).
// Use the helper methods to parse into typed values.
type FolderItem struct {
	ID       string `json:"id"`
	Linkcode string `json:"linkcode"`
	Name     string `json:"name"`
	Type     string `json:"type"`      // "0" (folder?)/other — semantics uncertain
	Path     string `json:"path"`      // parent path "/..."
	PID      string `json:"pid"`       // parent id
	Size     string `json:"size"`      // bytes (string-encoded)
	Mimetype string `json:"mimetype"`  // empty/null for folders, set for files
	Secure   string `json:"secure"`    // "0" | "1"
	Public   string `json:"public"`    // "0" | "1"
	FileType string `json:"file_type"` // fshare internal
	Created  string `json:"created"`   // unix ts string or ""
	Modified string `json:"modified"`  // unix ts string or ""
}

// IsFolder reports whether this item is a folder. Heuristic: folders have no
// mimetype set by fshare. Files always have a mimetype (video/x-matroska, etc).
func (f FolderItem) IsFolder() bool {
	return f.Mimetype == ""
}

// SizeBytes parses the string-encoded Size to int64. Returns 0 on error.
func (f FolderItem) SizeBytes() int64 {
	n, _ := parseInt64(f.Size)
	return n
}

// ModifiedTime parses the unix-timestamp string in Modified. Returns zero time
// if empty/unparseable.
func (f FolderItem) ModifiedTime() time.Time {
	n, err := parseInt64(f.Modified)
	if err != nil || n == 0 {
		return time.Time{}
	}
	return time.Unix(n, 0)
}

// DownloadRequest is the body sent to POST /api/session/download.
type DownloadRequest struct {
	Token    string `json:"token"`
	URL      string `json:"url"`      // "https://www.fshare.vn/file/<linkcode>"
	Password string `json:"password"` // "" if not password-protected
	Zipflag  int    `json:"zipflag"`
}

// DownloadResponse is the body returned from POST /api/session/download.
type DownloadResponse struct {
	Code     int    `json:"code,omitempty"`
	Location string `json:"location"`
	Msg      string `json:"msg,omitempty"`
}

// HealthResponse is the body returned when /api/user/get fails (session expired).
// Healthy responses return a full UserInfo object instead.
type HealthResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Session holds the active auth state for a Client.
type Session struct {
	Token       string
	SessionID   string
	Email       string
	AccountType string
	LastLoginAt time.Time
}

// Credentials stash the login email + password so the client can auto-relogin
// when the server returns code:201 (session expired) mid-call.
type Credentials struct {
	Email    string
	Password string
}
