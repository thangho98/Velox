package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSBrowse(t *testing.T) {
	t.Parallel()

	// Create a temp directory with some subdirs
	tmpDir := t.TempDir()
	subDir1 := filepath.Join(tmpDir, "movies")
	subDir2 := filepath.Join(tmpDir, "music")
	subDirHidden := filepath.Join(tmpDir, ".hidden")
	err := os.MkdirAll(subDir1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(subDir2, 0755)
	require.NoError(t, err)
	os.MkdirAll(subDirHidden, 0755)

	req := httptest.NewRequest(http.MethodGet, "/admin/fs/browse?path="+tmpDir, nil)
	w := httptest.NewRecorder()

	FSBrowse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Response is wrapped in {"data": ...}
	var wrapper map[string]fsBrowseResponse
	err = json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	resp, ok := wrapper["data"]
	require.True(t, ok)

	assert.Equal(t, tmpDir, resp.Current)
	assert.Equal(t, filepath.Dir(tmpDir), resp.Parent)
	require.Len(t, resp.Dirs, 2)

	// Check names are sorted
	names := make([]string, len(resp.Dirs))
	for i, d := range resp.Dirs {
		names[i] = d.Name
	}
	assert.Equal(t, []string{"movies", "music"}, names)

	// Hidden dir should not be included
	for _, d := range resp.Dirs {
		assert.NotEqual(t, ".hidden", d.Name)
	}
}

func TestFSBrowse_DefaultPath(t *testing.T) {
	t.Parallel()

	// When path is empty, defaults to "/"
	req := httptest.NewRequest(http.MethodGet, "/admin/fs/browse", nil)
	w := httptest.NewRecorder()

	FSBrowse(w, req)

	// Should try to browse "/" which likely fails
	// but we just verify it doesn't crash
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestFSBrowse_NonExistentPath(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/admin/fs/browse?path=/nonexistent/path", nil)
	w := httptest.NewRecorder()

	FSBrowse(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "cannot read directory")
}

func TestFSBrowse_FileInsteadOfDir(t *testing.T) {
	t.Parallel()

	// Create a temp file (not a directory)
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/admin/fs/browse?path="+tmpFile, nil)
	w := httptest.NewRecorder()

	FSBrowse(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFSBrowseResponse_JSON(t *testing.T) {
	t.Parallel()

	resp := fsBrowseResponse{
		Current: "/data",
		Parent:  "/",
		Dirs: []fsDirEntry{
			{Name: "movies", Path: "/data/movies"},
			{Name: "music", Path: "/data/music"},
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var got fsBrowseResponse
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, "/data", got.Current)
	assert.Equal(t, "/", got.Parent)
	require.Len(t, got.Dirs, 2)
	assert.Equal(t, "movies", got.Dirs[0].Name)
}
