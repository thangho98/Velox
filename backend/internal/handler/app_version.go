package handler

import (
	"encoding/json"
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
	if latest == nil {
		// Return 404 but as JSON so client can explicitly handle no updates
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no version found for platform"})
		return
	}

	downloadUrl := ""
	if platform == "android" {
		githubResp, err := http.Get("https://api.github.com/repos/thangho98/Velox/releases/latest")
		if err == nil {
			defer githubResp.Body.Close()
			if githubResp.StatusCode == 200 {
				var release struct {
					Assets []struct {
						Name               string `json:"name"`
						BrowserDownloadUrl string `json:"browser_download_url"`
					} `json:"assets"`
					HtmlUrl string `json:"html_url"`
				}
				if err := json.NewDecoder(githubResp.Body).Decode(&release); err == nil {
					downloadUrl = release.HtmlUrl
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
