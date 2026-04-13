package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thawng/velox/internal/repository"
)

type AppVersionHandler struct {
	repo *repository.AppVersionRepo
}

func NewAppVersionHandler(repo *repository.AppVersionRepo) *AppVersionHandler {
	return &AppVersionHandler{repo: repo}
}

func (h *AppVersionHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		http.Error(w, "missing platform parameter", http.StatusBadRequest)
		return
	}

	latest, err := h.repo.GetLatest(platform)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Create a dummy record if none exists to allow GitHub checking to still work
	if latest == nil {
		latest = &repository.AppVersion{
			Platform:    platform,
			VersionName: "0.1.0",
			VersionCode: 1,
			IsMandatory: false,
		}
	}

	downloadUrl := ""
	if platform == "android" {
		githubResp, err := http.Get("https://api.github.com/repos/thangho98/Velox/releases/latest")
		if err == nil {
			defer githubResp.Body.Close()
			if githubResp.StatusCode == 200 {
				var release struct {
					TagName string `json:"tag_name"`
					Assets  []struct {
						Name               string `json:"name"`
						BrowserDownloadUrl string `json:"browser_download_url"`
					} `json:"assets"`
					HtmlUrl string `json:"html_url"`
				}
				if err := json.NewDecoder(githubResp.Body).Decode(&release); err == nil {
					downloadUrl = release.HtmlUrl
					if release.TagName != "" {
						latest.VersionName = strings.TrimPrefix(release.TagName, "v")
						// simple heuristic for version_code: "0.1.5" -> 105
						parts := strings.Split(latest.VersionName, ".")
						if len(parts) >= 3 {
							var major, minor, patch int
							fmt.Sscanf(parts[0], "%d", &major)
							fmt.Sscanf(parts[1], "%d", &minor)
							fmt.Sscanf(parts[2], "%d", &patch)
							latest.VersionCode = major*10000 + minor*100 + patch
						}
					}
					for _, asset := range release.Assets {
						if strings.HasSuffix(asset.Name, ".apk") {
							downloadUrl = asset.BrowserDownloadUrl
							break
						}
					}
				}
			}
		}
	}

	resp := struct {
		Platform     string `json:"platform"`
		VersionName  string `json:"version_name"`
		VersionCode  int    `json:"version_code"`
		DownloadUrl  string `json:"download_url"`
		IsMandatory  bool   `json:"is_mandatory"`
		ReleaseNotes string `json:"release_notes"`
	}{
		Platform:     latest.Platform,
		VersionName:  latest.VersionName,
		VersionCode:  latest.VersionCode,
		DownloadUrl:  downloadUrl,
		IsMandatory:  latest.IsMandatory,
		ReleaseNotes: latest.ReleaseNotes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
