package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

type fsDirEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type fsBrowseResponse struct {
	Current string       `json:"current"`
	Parent  string       `json:"parent,omitempty"`
	Dirs    []fsDirEntry `json:"dirs"`
}

// FSBrowse returns the subdirectories at a given server path.
// Query param: ?path=/some/path (defaults to "/" if omitted).
// Admin-only — registered under RequireAdmin middleware.
func FSBrowse(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("path")
	if dir == "" {
		dir = "/"
	}
	dir = filepath.Clean(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		respondError(w, http.StatusBadRequest, "cannot read directory: "+err.Error())
		return
	}

	dirs := make([]fsDirEntry, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) > 0 && name[0] == '.' {
			continue // skip hidden dirs
		}
		dirs = append(dirs, fsDirEntry{
			Name: name,
			Path: filepath.Join(dir, name),
		})
	}
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})

	parent := filepath.Dir(dir)
	if parent == dir {
		parent = "" // already at filesystem root
	}

	respondJSON(w, http.StatusOK, fsBrowseResponse{
		Current: dir,
		Parent:  parent,
		Dirs:    dirs,
	})
}
