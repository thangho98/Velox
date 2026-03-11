package watcher

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/thawng/velox/internal/scanner"
)

// Watcher monitors library paths for file changes
type Watcher struct {
	watcher    *fsnotify.Watcher
	libraries  map[int64]string // library ID -> path
	onCreate   func(libraryID int64, path string)
	onRemove   func(libraryID int64, path string)
	onRename   func(libraryID int64, oldPath, newPath string)
	debouncers map[string]*time.Timer
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// Config for the watcher
type Config struct {
	Enabled bool
}

// New creates a new file watcher
func New(onCreate, onRemove func(libraryID int64, path string), onRename func(libraryID int64, oldPath, newPath string)) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		watcher:    fsWatcher,
		libraries:  make(map[int64]string),
		onCreate:   onCreate,
		onRemove:   onRemove,
		onRename:   onRename,
		debouncers: make(map[string]*time.Timer),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// AddLibrary adds a library path to watch
func (w *Watcher) AddLibrary(libraryID int64, path string) error {
	w.mu.Lock()
	w.libraries[libraryID] = path
	w.mu.Unlock()

	// Watch the root path
	if err := w.watcher.Add(path); err != nil {
		return err
	}

	// TODO: Recursively add subdirectories
	// For now, fsnotify doesn't support recursive watching on all platforms
	// We might need to walk and add each subdirectory

	return nil
}

// RemoveLibrary stops watching a library
func (w *Watcher) RemoveLibrary(libraryID int64) {
	w.mu.Lock()
	path, ok := w.libraries[libraryID]
	if ok {
		delete(w.libraries, libraryID)
		w.watcher.Remove(path)
	}
	w.mu.Unlock()
}

// Start begins watching for changes
func (w *Watcher) Start() {
	go w.run()
}

// Stop stops watching
func (w *Watcher) Stop() {
	w.cancel()
	w.watcher.Close()
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
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Find which library this path belongs to
	libraryID := w.findLibraryForPath(event.Name)
	if libraryID == 0 {
		return // Not in a watched library
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		w.debounceCreate(libraryID, event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		if w.onRemove != nil {
			w.onRemove(libraryID, event.Name)
		}
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		// fsnotify reports rename as two events: old path + new path
		// We need to track this - for now simplified
		if w.onRename != nil {
			// TODO: Implement proper rename tracking
		}
	}
}

func (w *Watcher) debounceCreate(libraryID int64, path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Cancel existing timer if any
	if timer, ok := w.debouncers[path]; ok {
		timer.Stop()
	}

	// Debounce for 5 seconds to avoid triggering while file is being written
	w.debouncers[path] = time.AfterFunc(5*time.Second, func() {
		w.mu.Lock()
		delete(w.debouncers, path)
		w.mu.Unlock()

		if w.onCreate != nil {
			w.onCreate(libraryID, path)
		}
	})
}

func (w *Watcher) findLibraryForPath(path string) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	for id, libPath := range w.libraries {
		if strings.HasPrefix(path, libPath) {
			return id
		}
	}
	return 0
}

// IsVideoFile checks if path is a video file (delegates to scanner package).
func IsVideoFile(path string) bool {
	return scanner.IsVideoFile(path)
}
