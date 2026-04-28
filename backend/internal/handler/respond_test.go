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
