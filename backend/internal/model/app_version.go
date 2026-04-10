package model

import "time"

type AppVersion struct {
	ID           int       `json:"id"`
	Platform     string    `json:"platform"`
	VersionName  string    `json:"version_name"`
	VersionCode  int       `json:"version_code"`
	IsMandatory  bool      `json:"is_mandatory"`
	ReleaseNotes string    `json:"release_notes"`
	CreatedAt    time.Time `json:"created_at"`
}
