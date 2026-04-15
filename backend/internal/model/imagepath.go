package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// ImagePath wraps a raw image path stored in the database. At rest it holds
// the same value we always have — `local://{type}/{id}/{file}` for uploads and
// the original TMDb relative path (`/abc123.jpg`) for external artwork.
//
// On JSON serialization it normalizes to an absolute-from-root URL that every
// client can use by simply prefixing the API base:
//
//	local://media/42/poster.jpg → /api/images/local/media/42/poster.jpg
//	/abc123.jpg                 → /api/images/tmdb/abc123.jpg
//	http(s)://...               → unchanged (external)
//	/api/...                    → unchanged (already normalized, idempotent)
//
// Size is not baked into the path. Clients pass `?width=N` when they need a
// specific size; the backend resolves that to the nearest TMDb size bucket
// and, for local uploads, serves the stored file (resize TBD).
type ImagePath string

// MarshalJSON emits the normalized URL form.
func (p ImagePath) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Normalize())
}

// UnmarshalJSON accepts either the raw form (local://, /abc.jpg) or the
// already-normalized form. The normalizer is idempotent so round-tripping is
// safe.
func (p *ImagePath) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = ImagePath(s)
	return nil
}

// Normalize returns the client-facing URL form of the stored path, fallback without specific type.
func (p ImagePath) Normalize() string {
	return p.NormalizeAs("")
}

func (p ImagePath) NormalizeAs(kind string) string {
	s := string(p)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	if strings.HasPrefix(s, "/api/") {
		return s
	}
	if strings.HasPrefix(s, "local://") {
		return "/api/images/local/" + strings.TrimPrefix(s, "local://")
	}
	if strings.HasPrefix(s, "/") {
		if kind != "" {
			return "/api/images/tmdb/" + kind + s
		}
		// Untyped caller: no image kind known, so we can't build a valid
		// /api/images/tmdb/{type}/{path} route. Return the raw path unchanged
		// rather than synthesizing a 404-ing endpoint. All production callers
		// should use a typed path wrapper (PosterPath/BackdropPath/ProfilePath/…).
		return s
	}
	return s
}

// PosterPath wraps a raw image path and normalizes as a poster type.
type PosterPath string

func (p PosterPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(ImagePath(p).NormalizeAs("poster"))
}
func (p *PosterPath) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = PosterPath(s)
	return nil
}
func (p PosterPath) Normalize() string {
	return ImagePath(p).NormalizeAs("poster")
}
func (p PosterPath) Raw() string {
	return string(p)
}
func (p *PosterPath) Scan(src interface{}) error {
	var ip ImagePath
	if err := ip.Scan(src); err != nil {
		return err
	}
	*p = PosterPath(ip.Raw())
	return nil
}
func (p PosterPath) Value() (driver.Value, error) {
	return string(p), nil
}

// BackdropPath wraps a raw image path and normalizes as a backdrop type.
type BackdropPath string

func (p BackdropPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(ImagePath(p).NormalizeAs("backdrop"))
}
func (p *BackdropPath) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = BackdropPath(s)
	return nil
}
func (p BackdropPath) Normalize() string {
	return ImagePath(p).NormalizeAs("backdrop")
}
func (p BackdropPath) Raw() string {
	return string(p)
}
func (p *BackdropPath) Scan(src interface{}) error {
	var ip ImagePath
	if err := ip.Scan(src); err != nil {
		return err
	}
	*p = BackdropPath(ip.Raw())
	return nil
}
func (p BackdropPath) Value() (driver.Value, error) {
	return string(p), nil
}

// StillPath wraps a raw image path and normalizes as a still type.
type StillPath string

func (p StillPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(ImagePath(p).NormalizeAs("still"))
}
func (p *StillPath) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = StillPath(s)
	return nil
}
func (p StillPath) Normalize() string {
	return ImagePath(p).NormalizeAs("still")
}
func (p StillPath) Raw() string {
	return string(p)
}
func (p *StillPath) Scan(src interface{}) error {
	var ip ImagePath
	if err := ip.Scan(src); err != nil {
		return err
	}
	*p = StillPath(ip.Raw())
	return nil
}
func (p StillPath) Value() (driver.Value, error) {
	return string(p), nil
}

// ProfilePath wraps a raw image path (actor/crew headshot) and normalizes as
// a profile type.
type ProfilePath string

func (p ProfilePath) MarshalJSON() ([]byte, error) {
	return json.Marshal(ImagePath(p).NormalizeAs("profile"))
}
func (p *ProfilePath) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = ProfilePath(s)
	return nil
}
func (p ProfilePath) Normalize() string {
	return ImagePath(p).NormalizeAs("profile")
}
func (p ProfilePath) Raw() string {
	return string(p)
}
func (p *ProfilePath) Scan(src interface{}) error {
	var ip ImagePath
	if err := ip.Scan(src); err != nil {
		return err
	}
	*p = ProfilePath(ip.Raw())
	return nil
}
func (p ProfilePath) Value() (driver.Value, error) {
	return string(p), nil
}

// LogoPath wraps a raw image path and normalizes as a logo type.
type LogoPath string

func (p LogoPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(ImagePath(p).NormalizeAs("logo"))
}
func (p *LogoPath) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = LogoPath(s)
	return nil
}
func (p LogoPath) Normalize() string {
	return ImagePath(p).NormalizeAs("logo")
}
func (p LogoPath) Raw() string {
	return string(p)
}
func (p *LogoPath) Scan(src interface{}) error {
	var ip ImagePath
	if err := ip.Scan(src); err != nil {
		return err
	}
	*p = LogoPath(ip.Raw())
	return nil
}
func (p LogoPath) Value() (driver.Value, error) {
	return string(p), nil
}

// Raw returns the underlying stored value (for DB writes, comparisons, etc.).
func (p ImagePath) Raw() string {
	return string(p)
}

// Scan implements sql.Scanner so repositories can scan directly into
// ImagePath-typed struct fields.
func (p *ImagePath) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		*p = ""
	case string:
		*p = ImagePath(v)
	case []byte:
		*p = ImagePath(v)
	default:
		return fmt.Errorf("ImagePath.Scan: unsupported type %T", src)
	}
	return nil
}

// Value implements driver.Valuer so repositories can pass ImagePath directly
// to exec/query placeholders.
func (p ImagePath) Value() (driver.Value, error) {
	return string(p), nil
}
