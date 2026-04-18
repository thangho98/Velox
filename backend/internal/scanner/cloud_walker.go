package scanner

import (
	"context"
	"errors"
	"fmt"
	neturl "net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/thawng/velox/internal/cloudstorage"
)

// CloudPathPrefix is the "scheme://id" delimiter used in media_files.path to
// encode cloud-sourced media (e.g. "fshare://XZWCPAZV3J71"). Scanner writes
// this shape; stream handler parses it back to (provider_type, native_id).
const CloudPathPrefix = "://"

// CloudFileCallback is invoked for every video file discovered during a cloud
// walk. Implementations typically insert rows into media + media_files. The
// callback receives the provider-scheme path ready for direct persistence,
// along with the relative path inside the folder for accurate metadata scraping.
type CloudFileCallback func(ctx context.Context, item cloudstorage.Item, storagePath string, relativePath string) error

// CloudWalker performs a breadth-first traversal of a cloud provider's folder
// tree, invoking the callback for each video file. It is provider-agnostic —
// swap fshare for GoogleDrive by passing a different cloudstorage.Provider.
type CloudWalker struct {
	provider cloudstorage.Provider
	onFile   CloudFileCallback
	// onFolderErr is called for non-fatal errors (single folder lookup fails);
	// the walk continues on other folders so one bad subtree doesn't abort scan.
	onFolderErr func(folderID string, err error)
}

// NewCloudWalker builds a CloudWalker. onFolderErr may be nil (errors are
// silently dropped but walk continues).
func NewCloudWalker(provider cloudstorage.Provider, onFile CloudFileCallback, onFolderErr func(string, error)) *CloudWalker {
	return &CloudWalker{
		provider:    provider,
		onFile:      onFile,
		onFolderErr: onFolderErr,
	}
}

type cloudFolderNode struct {
	ID   string
	Path string
}

// Walk traverses rootFolderID and its subfolders. Cancellation via ctx stops
// the walk cleanly at the next folder boundary.
func (w *CloudWalker) Walk(ctx context.Context, rootFolderID string) error {
	var (
		queue = []cloudFolderNode{{ID: rootFolderID, Path: ""}}
		seen  = map[string]bool{}
		mu    sync.Mutex
	)

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		current := queue[0]
		queue = queue[1:]

		mu.Lock()
		if seen[current.ID] {
			mu.Unlock()
			continue
		}
		seen[current.ID] = true
		mu.Unlock()

		items, err := w.provider.ListFolder(ctx, current.ID)
		if err != nil {
			// Session-level errors abort the entire walk; per-folder errors
			// (e.g. not-found) fall through to the optional callback.
			if errors.Is(err, cloudstorage.ErrSessionExpired) ||
				errors.Is(err, cloudstorage.ErrInvalidCredentials) ||
				errors.Is(err, cloudstorage.ErrRateLimit) {
				return fmt.Errorf("cloud walk aborted: %w", err)
			}
			if w.onFolderErr != nil {
				w.onFolderErr(current.ID, err)
			}
			continue
		}

		for _, item := range items {
			relPath := item.Name
			if current.Path != "" {
				relPath = current.Path + "/" + item.Name
			}

			if item.IsFolder {
				queue = append(queue, cloudFolderNode{ID: item.ID, Path: relPath})
				continue
			}
			if !isCloudVideo(item) {
				continue
			}
			// Embed relative path via query param so we never lose original filename parsing ability
			storagePath := w.provider.Type() + CloudPathPrefix + item.ID + "?name=" + neturl.QueryEscape(relPath)
			if err := w.onFile(ctx, item, storagePath, relPath); err != nil {
				if w.onFolderErr != nil {
					w.onFolderErr(current.ID, fmt.Errorf("file %s: %w", item.Name, err))
				}
				continue
			}
		}
	}
	return nil
}

// isCloudVideo reports whether an item is a playable video. Cloud providers
// sometimes omit mimetype for valid video files, so fall back to extension.
func isCloudVideo(item cloudstorage.Item) bool {
	if strings.HasPrefix(item.Mimetype, "video/") {
		return true
	}
	return VideoExtensions[strings.ToLower(filepath.Ext(item.Name))]
}

// ParseCloudPath splits "fshare://abc?name=..." into ("fshare", "abc").
// Returns ok == false when the path isn't a cloud-scheme path.
func ParseCloudPath(path string) (providerType, nativeID string, ok bool) {
	idx := strings.Index(path, CloudPathPrefix)
	if idx <= 0 {
		return "", "", false
	}
	remaining := path[idx+len(CloudPathPrefix):]
	if remaining == "" {
		return "", "", false
	}

	nativeID = remaining
	if qIdx := strings.IndexByte(remaining, '?'); qIdx >= 0 {
		nativeID = remaining[:qIdx]
	}
	if nativeID == "" {
		return "", "", false
	}

	return path[:idx], nativeID, true
}

// IsCloudPath reports whether path was emitted by CloudWalker.
func IsCloudPath(path string) bool {
	_, _, ok := ParseCloudPath(path)
	return ok
}
