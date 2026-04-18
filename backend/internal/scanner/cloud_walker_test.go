package scanner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/thawng/velox/internal/cloudstorage"
)

// fakeProvider implements cloudstorage.Provider for Walker tests.
type fakeProvider struct {
	typeName string
	folders  map[string][]cloudstorage.Item
	calls    atomic.Int32
	errOn    map[string]error // override: folderID → error
}

func (p *fakeProvider) Type() string                          { return p.typeName }
func (p *fakeProvider) ParseFolderURL(string) (string, error) { return "", nil }
func (p *fakeProvider) ParseFileURL(string) (string, error)   { return "", nil }
func (p *fakeProvider) GetDownloadURL(context.Context, string) (string, error) {
	return "", nil
}
func (p *fakeProvider) GetAccountInfo(context.Context) (*cloudstorage.AccountInfo, error) {
	return nil, nil
}
func (p *fakeProvider) CheckHealth(context.Context) error  { return nil }
func (p *fakeProvider) Session() *cloudstorage.Credentials { return nil }

func (p *fakeProvider) ListFolder(_ context.Context, id string) ([]cloudstorage.Item, error) {
	p.calls.Add(1)
	if err, ok := p.errOn[id]; ok {
		return nil, err
	}
	return p.folders[id], nil
}

func TestCloudWalker_BFS_ProcessesAllVideos(t *testing.T) {
	provider := &fakeProvider{
		typeName: "fshare",
		folders: map[string][]cloudstorage.Item{
			"root": {
				{ID: "subA", Name: "Movies", IsFolder: true},
				{ID: "subB", Name: "Shows", IsFolder: true},
				{ID: "notes.txt", Name: "notes.txt", Mimetype: "text/plain"},
			},
			"subA": {
				{ID: "m1", Name: "Avatar.2009.mkv", Mimetype: "video/x-matroska", Size: 5_000_000_000},
			},
			"subB": {
				{ID: "s1", Name: "Show.S01E01.mp4", Size: 1_000_000_000}, // no mimetype → extension fallback
			},
		},
	}

	var (
		mu   sync.Mutex
		hits []string
	)
	walker := NewCloudWalker(provider, func(_ context.Context, item cloudstorage.Item, storagePath string, relativePath string) error {
		mu.Lock()
		defer mu.Unlock()
		hits = append(hits, storagePath)
		return nil
	}, nil)

	if err := walker.Walk(context.Background(), "root"); err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(hits) != 2 {
		t.Errorf("visited = %v; want 2 files", hits)
	}
	// Storage paths use provider-scheme prefix.
	for _, p := range hits {
		if provType, _, ok := ParseCloudPath(p); !ok || provType != "fshare" {
			t.Errorf("path %q missing fshare:// prefix", p)
		}
	}
}

func TestCloudWalker_SkipsNonVideoByMime(t *testing.T) {
	provider := &fakeProvider{
		typeName: "fshare",
		folders: map[string][]cloudstorage.Item{
			"root": {
				{ID: "a", Name: "poster.jpg", Mimetype: "image/jpeg"},
				{ID: "b", Name: "audio.mp3", Mimetype: "audio/mpeg"},
				{ID: "c", Name: "video.mkv", Mimetype: "video/x-matroska"},
			},
		},
	}
	var count int
	walker := NewCloudWalker(provider, func(context.Context, cloudstorage.Item, string, string) error {
		count++
		return nil
	}, nil)
	_ = walker.Walk(context.Background(), "root")
	if count != 1 {
		t.Errorf("visited = %d; want 1", count)
	}
}

func TestCloudWalker_CycleProtection(t *testing.T) {
	provider := &fakeProvider{
		typeName: "fshare",
		folders: map[string][]cloudstorage.Item{
			"root": {{ID: "loop", Name: "Loop", IsFolder: true}},
			"loop": {{ID: "root", Name: "Back", IsFolder: true}, {ID: "v1", Name: "video.mkv"}},
		},
	}
	var count int
	walker := NewCloudWalker(provider, func(context.Context, cloudstorage.Item, string, string) error {
		count++
		return nil
	}, nil)
	if err := walker.Walk(context.Background(), "root"); err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if count != 1 {
		t.Errorf("visited = %d; cycle protection failed", count)
	}
	// 2 unique folders visited, not more.
	if n := provider.calls.Load(); n != 2 {
		t.Errorf("ListFolder called %d times; want 2 (root + loop)", n)
	}
}

func TestCloudWalker_PerFolderErrorContinues(t *testing.T) {
	provider := &fakeProvider{
		typeName: "fshare",
		folders: map[string][]cloudstorage.Item{
			"root": {{ID: "bad", Name: "Bad", IsFolder: true}, {ID: "good", Name: "Good", IsFolder: true}},
			"good": {{ID: "v1", Name: "video.mkv"}},
		},
		errOn: map[string]error{"bad": cloudstorage.ErrFileNotFound},
	}
	var (
		visited int
		errors  []string
	)
	walker := NewCloudWalker(provider,
		func(context.Context, cloudstorage.Item, string, string) error { visited++; return nil },
		func(folder string, err error) { errors = append(errors, folder) },
	)
	if err := walker.Walk(context.Background(), "root"); err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if visited != 1 {
		t.Errorf("visited = %d; want 1", visited)
	}
	if len(errors) != 1 || errors[0] != "bad" {
		t.Errorf("errors = %v; want [bad]", errors)
	}
}

func TestCloudWalker_SessionExpiredAborts(t *testing.T) {
	provider := &fakeProvider{
		typeName: "fshare",
		folders:  map[string][]cloudstorage.Item{},
		errOn:    map[string]error{"root": cloudstorage.ErrSessionExpired},
	}
	walker := NewCloudWalker(provider,
		func(context.Context, cloudstorage.Item, string, string) error { return nil },
		nil,
	)
	err := walker.Walk(context.Background(), "root")
	if !errors.Is(err, cloudstorage.ErrSessionExpired) {
		t.Errorf("err = %v; want ErrSessionExpired", err)
	}
}

func TestParseCloudPath(t *testing.T) {
	cases := []struct {
		in           string
		providerType string
		nativeID     string
		ok           bool
	}{
		{"fshare://abc", "fshare", "abc", true},
		{"gdrive://1a2b3c", "gdrive", "1a2b3c", true},
		{"/mnt/data/foo.mkv", "", "", false},
		{"://only", "", "", false},
		{"fshare://", "", "", false},
		{"no-separator", "", "", false},
	}
	for _, tc := range cases {
		gotType, gotID, ok := ParseCloudPath(tc.in)
		if ok != tc.ok {
			t.Errorf("%q: ok=%v; want %v", tc.in, ok, tc.ok)
		}
		if gotType != tc.providerType || gotID != tc.nativeID {
			t.Errorf("%q: got (%q,%q); want (%q,%q)", tc.in, gotType, gotID, tc.providerType, tc.nativeID)
		}
	}
}
