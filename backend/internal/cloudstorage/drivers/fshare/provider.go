package fshare

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/thawng/velox/internal/cloudstorage"
	pkgfshare "github.com/thawng/velox/pkg/fshare"
)

// provider wraps a pkg/fshare.Client, exposing the cloudstorage.Provider
// interface. Session state is owned by the client; provider only mirrors the
// snapshot so callers can detect auto-relogins.
type provider struct {
	client *pkgfshare.Client

	mu    sync.Mutex
	creds cloudstorage.Credentials // snapshot of last known credentials
}

func (p *provider) Type() string { return "fshare" }

func (p *provider) ParseFolderURL(rawURL string) (string, error) { return parseFolderURL(rawURL) }
func (p *provider) ParseFileURL(rawURL string) (string, error)   { return parseFileURL(rawURL) }

func (p *provider) ListFolder(ctx context.Context, folderID string) ([]cloudstorage.Item, error) {
	items, err := p.client.ListFolder(ctx, folderID)
	if err != nil {
		return nil, mapError(err)
	}
	p.refreshSnapshot()
	return convertItems(items), nil
}

func (p *provider) GetDownloadURL(ctx context.Context, fileID string) (string, error) {
	url, err := p.client.GetDirectLink(ctx, fileID)
	if err != nil {
		return "", mapError(err)
	}
	p.refreshSnapshot()
	return url, nil
}

func (p *provider) GetAccountInfo(ctx context.Context) (*cloudstorage.AccountInfo, error) {
	info, err := p.client.GetUserInfo(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	p.refreshSnapshot()

	quota, _ := strconv.ParseInt(info.Webspace, 10, 64)
	var expiry time.Time
	if ts, err := strconv.ParseInt(info.ExpireVIP, 10, 64); err == nil && ts > 0 {
		expiry = time.Unix(ts, 0)
	}
	return &cloudstorage.AccountInfo{
		Email:       info.Email,
		DisplayName: info.Name,
		AccountType: info.AccountType,
		QuotaBytes:  quota,
		ExpiresAt:   expiry,
	}, nil
}

func (p *provider) CheckHealth(ctx context.Context) error {
	if err := p.client.CheckSession(ctx); err != nil {
		return mapError(err)
	}
	p.refreshSnapshot()
	return nil
}

// Session returns a snapshot of the current credentials state so callers can
// compare against the pre-call value and persist when a silent relogin happened.
func (p *provider) Session() *cloudstorage.Credentials {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := p.creds
	return &out
}

// refreshSnapshot copies the client's current session into the provider's
// snapshot. Called after every API call so Session() observes relogins.
func (p *provider) refreshSnapshot() {
	sess := p.client.Session()
	if sess == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if sess.Token != p.creds.AccessToken || sess.SessionID != p.creds.SessionID {
		p.creds.AccessToken = sess.Token
		p.creds.SessionID = sess.SessionID
		p.creds.ExpiresAt = time.Now().Add(sessionLifetime)
	}
	// Email/AccountType get populated by GetAccountInfo; don't overwrite here.
}

// convertItems translates pkg/fshare.FolderItem into cloudstorage.Item.
func convertItems(src []pkgfshare.FolderItem) []cloudstorage.Item {
	out := make([]cloudstorage.Item, 0, len(src))
	for _, it := range src {
		canonical := fileURL(it.Linkcode)
		if it.IsFolder() {
			canonical = folderURL(it.Linkcode)
		}
		out = append(out, cloudstorage.Item{
			ID:       it.Linkcode,
			Name:     it.Name,
			IsFolder: it.IsFolder(),
			Size:     it.SizeBytes(),
			Mimetype: it.Mimetype,
			Modified: it.ModifiedTime(),
			URL:      canonical,
		})
	}
	return out
}
