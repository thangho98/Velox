package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/thawng/velox/internal/scanner"
)

// Watcher monitors library paths for file changes and triggers scan callbacks.
type Watcher struct {
	watcher   *fsnotify.Watcher
	libraries map[int64][]string // library ID -> root paths

	onFileCreated func(libraryID int64, path string)
	onFileRemoved func(libraryID int64, path string)

	// Debounce CREATE events — files may still be copying when the event fires.
	debouncers map[string]*time.Timer
	debounceD  time.Duration

	// Rename tracking: fsnotify fires RENAME on old path, then CREATE on new.
	// Buffer the RENAME for a short window to pair with the following CREATE.
	pendingRenames map[string]int64 // old path -> libraryID
	renameTTL      time.Duration

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new file watcher.
//   - onFileCreated is called (after debounce) when a new video file appears.
//   - onFileRemoved is called when a video file is deleted.
func New(
	onFileCreated func(libraryID int64, path string),
	onFileRemoved func(libraryID int64, path string),
) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		watcher:        fsWatcher,
		libraries:      make(map[int64][]string),
		onFileCreated:  onFileCreated,
		onFileRemoved:  onFileRemoved,
		debouncers:     make(map[string]*time.Timer),
		debounceD:      5 * time.Second,
		pendingRenames: make(map[string]int64),
		renameTTL:      2 * time.Second,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// AddLibrary adds all directories under the library's root paths to the watch list.
func (w *Watcher) AddLibrary(libraryID int64, paths []string) error {
	w.mu.Lock()
	w.libraries[libraryID] = paths
	w.mu.Unlock()

	for _, root := range paths {
		if err := w.watchRecursive(root); err != nil {
			return err
		}
	}
	return nil
}

// RemoveLibrary stops watching a library's paths.
func (w *Watcher) RemoveLibrary(libraryID int64) {
	w.mu.Lock()
	paths, ok := w.libraries[libraryID]
	if ok {
		delete(w.libraries, libraryID)
	}
	w.mu.Unlock()

	if ok {
		for _, root := range paths {
			// Best-effort removal — fsnotify ignores unknown paths.
			_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					w.watcher.Remove(path)
				}
				return nil
			})
		}
	}
}

// Start begins the event loop in a background goroutine.
func (w *Watcher) Start() {
	go w.run()
}

// Stop gracefully shuts down the watcher.
func (w *Watcher) Stop() {
	w.cancel()
	w.watcher.Close()
}

// watchRecursive adds a directory and all its subdirectories to fsnotify.
func (w *Watcher) watchRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}
		if d.IsDir() {
			if err := w.watcher.Add(path); err != nil {
				log.Printf("watcher: failed to watch %s: %v", path, err)
			}
		}
		return nil
	})
}

func (w *Watcher) run() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	path := event.Name

	switch {
	case event.Op&fsnotify.Create != 0:
		// New directory → watch it recursively for future events.
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			w.watchRecursive(path)
			return
		}

		if !scanner.IsVideoFile(path) {
			return
		}

		libraryID := w.findLibraryForPath(path)
		if libraryID == 0 {
			return
		}

		// Check if this CREATE is the second half of a rename.
		// The pipeline's fingerprint check already handles renames correctly,
		// so we just treat it as a normal "new file" event — the pipeline
		// will detect the fingerprint match and update the path.
		w.clearPendingRename(path)
		w.debounceCreate(libraryID, path)

	case event.Op&fsnotify.Remove != 0:
		if !scanner.IsVideoFile(path) {
			return
		}

		libraryID := w.findLibraryForPath(path)
		if libraryID == 0 {
			return
		}

		if w.onFileRemoved != nil {
			w.onFileRemoved(libraryID, path)
		}

	case event.Op&fsnotify.Rename != 0:
		if !scanner.IsVideoFile(path) {
			return
		}

		libraryID := w.findLibraryForPath(path)
		if libraryID == 0 {
			return
		}

		// Buffer the rename — a CREATE for the new path should follow shortly.
		// If no CREATE arrives within renameTTL, treat it as a removal.
		w.mu.Lock()
		w.pendingRenames[path] = libraryID
		w.mu.Unlock()

		time.AfterFunc(w.renameTTL, func() {
			w.mu.Lock()
			libID, still := w.pendingRenames[path]
			if still {
				delete(w.pendingRenames, path)
			}
			w.mu.Unlock()

			if still && w.onFileRemoved != nil {
				// No matching CREATE arrived — the file was truly removed/renamed out.
				w.onFileRemoved(libID, path)
			}
		})

	case event.Op&fsnotify.Write != 0:
		// Ignore write events — files are indexed on CREATE.
	}
}

func (w *Watcher) debounceCreate(libraryID int64, path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if timer, ok := w.debouncers[path]; ok {
		timer.Stop()
	}

	w.debouncers[path] = time.AfterFunc(w.debounceD, func() {
		w.mu.Lock()
		delete(w.debouncers, path)
		w.mu.Unlock()

		if w.onFileCreated != nil {
			w.onFileCreated(libraryID, path)
		}
	})
}

func (w *Watcher) clearPendingRename(path string) {
	w.mu.Lock()
	delete(w.pendingRenames, path)
	w.mu.Unlock()
}

func (w *Watcher) findLibraryForPath(path string) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	for id, roots := range w.libraries {
		for _, root := range roots {
			if strings.HasPrefix(path, root) {
				return id
			}
		}
	}
	return 0
}
