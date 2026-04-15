package repository

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thawng/velox/internal/model"
)

// BrowseFolderItem represents a subfolder in browse results
type BrowseFolderItem struct {
	Name       string           `json:"name"`
	Path       string           `json:"path"`             // library-relative path for navigation
	MediaCount int              `json:"media_count"`      // number of media files under this folder
	Poster     model.PosterPath `json:"poster,omitempty"` // poster from first media in folder (Emby-style)
}

// BrowseResult represents the result of a folder browse operation.
// Includes both subfolders and media items directly in the current folder.
type BrowseResult struct {
	LibraryID int64                 `json:"library_id"`
	Path      string                `json:"path"`   // current library-relative path
	Parent    string                `json:"parent"` // parent relative path, "" if root
	Folders   []BrowseFolderItem    `json:"folders"`
	Media     []model.MediaListItem `json:"media"`
}

// BrowseFolders returns subfolders + media at a given path within a library.
// absDir is the resolved absolute directory path on disk.
// relativePath is the library-relative path for the response.
func (r *MediaFileRepo) BrowseFolders(ctx context.Context, libraryID int64, absDir, relativePath string) (*BrowseResult, error) {
	// Ensure absDir ends without trailing slash for consistent LIKE matching
	absDir = strings.TrimRight(absDir, "/")
	prefix := absDir + "/"

	result := &BrowseResult{
		LibraryID: libraryID,
		Path:      relativePath,
		Folders:   []BrowseFolderItem{},
		Media:     []model.MediaListItem{},
	}

	// Compute parent path
	if relativePath != "" {
		parent := filepath.Dir(relativePath)
		if parent == "." {
			parent = ""
		}
		result.Parent = parent
	}

	// Step 1: Get all file paths under this directory to extract subdirectories
	rows, err := r.db.QueryContext(ctx, `
		SELECT mf.file_path
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ?
		  AND mf.file_path LIKE ? || '%'`,
		libraryID, prefix)
	if err != nil {
		return nil, fmt.Errorf("browse listing files: %w", err)
	}
	defer rows.Close()

	// Extract immediate subdirectory names and count media per subdir
	dirCounts := map[string]int{}
	for rows.Next() {
		var fp string
		if err := rows.Scan(&fp); err != nil {
			return nil, fmt.Errorf("browse scanning path: %w", err)
		}
		// Strip prefix to get relative remainder: "SubFolder/file.mkv" or "file.mkv"
		rel := strings.TrimPrefix(fp, prefix)
		if slashIdx := strings.Index(rel, "/"); slashIdx > 0 {
			// Has subdirectory — count it
			dirName := rel[:slashIdx]
			dirCounts[dirName]++
		}
		// Files directly in this folder (no slash) are handled in Step 2
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("browse iterating paths: %w", err)
	}

	// Sort folder names alphabetically
	dirNames := make([]string, 0, len(dirCounts))
	for name := range dirCounts {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	for _, name := range dirNames {
		folderRelPath := name
		if relativePath != "" {
			folderRelPath = relativePath + "/" + name
		}
		// Fetch poster from first media in this subfolder
		subPrefix := prefix + name + "/"
		var poster sql.NullString
		_ = r.db.QueryRowContext(ctx, `
			SELECT m.poster_path FROM media_files mf
			JOIN media m ON m.id = mf.media_id
			WHERE m.library_id = ? AND mf.file_path LIKE ? || '%' AND m.poster_path != ''
			ORDER BY m.sort_title LIMIT 1`,
			libraryID, subPrefix).Scan(&poster)

		result.Folders = append(result.Folders, BrowseFolderItem{
			Name:       name,
			Path:       folderRelPath,
			MediaCount: dirCounts[name],
			Poster:     model.PosterPath(poster.String),
		})
	}

	// Step 2: Get media items directly in this folder (not in subdirectories)
	mediaRows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
		       m.release_date, m.rating, m.overview,
		       GROUP_CONCAT(DISTINCT g.name) as genre_names,
		       COALESCE(e.series_id, 0) as series_id
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN episodes e ON e.media_id = m.id
		WHERE m.library_id = ?
		  AND mf.file_path LIKE ? || '%'
		  AND mf.file_path NOT LIKE ? || '%/%'
		GROUP BY m.id
		ORDER BY m.sort_title`,
		libraryID, prefix, prefix)
	if err != nil {
		return nil, fmt.Errorf("browse listing media: %w", err)
	}
	defer mediaRows.Close()

	for mediaRows.Next() {
		var item model.MediaListItem
		var genreNames sql.NullString
		if err := mediaRows.Scan(
			&item.ID, &item.Title, &item.SortTitle, &item.PosterPath, &item.MediaType,
			&item.ReleaseDate, &item.Rating, &item.Overview,
			&genreNames, &item.SeriesID,
		); err != nil {
			return nil, fmt.Errorf("browse scanning media: %w", err)
		}
		if genreNames.Valid && genreNames.String != "" {
			item.Genres = strings.Split(genreNames.String, ",")
		}
		result.Media = append(result.Media, item)
	}

	return result, mediaRows.Err()
}
