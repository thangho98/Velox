package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/model"
)

func setupStorageProviderTestDB(t *testing.T) (*sql.DB, *StorageProviderRepo) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)

	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			is_admin INTEGER DEFAULT 0
		);
		INSERT INTO users (id, username, display_name, password_hash) VALUES (1, 'u1', 'U1', 'x');

		CREATE TABLE storage_providers (
			id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id               INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			provider_type         TEXT NOT NULL,
			display_name          TEXT NOT NULL,
			account_email         TEXT NOT NULL,
			credentials_encrypted BLOB NOT NULL,
			account_type          TEXT DEFAULT NULL,
			quota_bytes           INTEGER DEFAULT NULL,
			used_bytes            INTEGER DEFAULT NULL,
			account_expires_at    DATETIME DEFAULT NULL,
			token_expires_at      DATETIME DEFAULT NULL,
			last_refresh_at       DATETIME DEFAULT NULL,
			last_check_at         DATETIME DEFAULT NULL,
			last_error            TEXT DEFAULT NULL,
			created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, provider_type, account_email)
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db, NewStorageProviderRepo(db)
}

func TestStorageProviderRepo_CreateAndGet(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	ctx := context.Background()

	expiry := time.Now().Add(25 * time.Minute)
	vip := "Vip"
	sp := &model.StorageProvider{
		UserID:               1,
		ProviderType:         "fshare",
		DisplayName:          "My Fshare",
		AccountEmail:         "user@example.com",
		CredentialsEncrypted: []byte{0x01, 0x02, 0x03},
		AccountType:          &vip,
		TokenExpiresAt:       &expiry,
	}
	if err := repo.Create(ctx, sp); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sp.ID == 0 {
		t.Errorf("ID not populated after Create")
	}
	if sp.CreatedAt.IsZero() {
		t.Errorf("CreatedAt not populated")
	}

	fetched, err := repo.GetByID(ctx, sp.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.AccountEmail != "user@example.com" {
		t.Errorf("email = %q", fetched.AccountEmail)
	}
	if fetched.AccountType == nil || *fetched.AccountType != "Vip" {
		t.Errorf("account_type = %v", fetched.AccountType)
	}
	if len(fetched.CredentialsEncrypted) != 3 {
		t.Errorf("credentials blob size = %d; want 3", len(fetched.CredentialsEncrypted))
	}
}

func TestStorageProviderRepo_GetByID_NotFound(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	_, err := repo.GetByID(context.Background(), 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v; want ErrNotFound", err)
	}
}

func TestStorageProviderRepo_GetByAccount(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	ctx := context.Background()

	sp := &model.StorageProvider{
		UserID:               1,
		ProviderType:         "fshare",
		DisplayName:          "A",
		AccountEmail:         "a@example.com",
		CredentialsEncrypted: []byte{0x00},
	}
	_ = repo.Create(ctx, sp)

	fetched, err := repo.GetByAccount(ctx, 1, "fshare", "a@example.com")
	if err != nil {
		t.Fatalf("GetByAccount: %v", err)
	}
	if fetched.ID != sp.ID {
		t.Errorf("id mismatch")
	}

	_, err = repo.GetByAccount(ctx, 1, "fshare", "missing@example.com")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v; want ErrNotFound", err)
	}
}

func TestStorageProviderRepo_Update(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	ctx := context.Background()

	sp := &model.StorageProvider{
		UserID:               1,
		ProviderType:         "fshare",
		DisplayName:          "Old Name",
		AccountEmail:         "u@example.com",
		CredentialsEncrypted: []byte{0x01},
	}
	if err := repo.Create(ctx, sp); err != nil {
		t.Fatalf("Create: %v", err)
	}

	newExpiry := time.Now().Add(30 * time.Minute)
	sp.DisplayName = "New Name"
	sp.CredentialsEncrypted = []byte{0x02, 0x03}
	sp.TokenExpiresAt = &newExpiry

	if err := repo.Update(ctx, sp); err != nil {
		t.Fatalf("Update: %v", err)
	}

	fetched, _ := repo.GetByID(ctx, sp.ID)
	if fetched.DisplayName != "New Name" {
		t.Errorf("DisplayName = %q", fetched.DisplayName)
	}
	if len(fetched.CredentialsEncrypted) != 2 {
		t.Errorf("creds blob size = %d; want 2", len(fetched.CredentialsEncrypted))
	}
}

func TestStorageProviderRepo_Update_NotFound(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	sp := &model.StorageProvider{ID: 9999, DisplayName: "ghost"}
	err := repo.Update(context.Background(), sp)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v; want ErrNotFound", err)
	}
}

func TestStorageProviderRepo_ListExpiringSoon(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	ctx := context.Background()

	now := time.Now()
	mkProvider := func(email string, expiresAt *time.Time) *model.StorageProvider {
		return &model.StorageProvider{
			UserID:               1,
			ProviderType:         "fshare",
			DisplayName:          email,
			AccountEmail:         email,
			CredentialsEncrypted: []byte{0x00},
			TokenExpiresAt:       expiresAt,
		}
	}

	soonExpiry := now.Add(2 * time.Minute)
	healthyExpiry := now.Add(1 * time.Hour)

	_ = repo.Create(ctx, mkProvider("soon@x.com", &soonExpiry))
	_ = repo.Create(ctx, mkProvider("healthy@x.com", &healthyExpiry))
	_ = repo.Create(ctx, mkProvider("no-expiry@x.com", nil))

	expiring, err := repo.ListExpiringSoon(ctx, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("ListExpiringSoon: %v", err)
	}
	if len(expiring) != 1 || expiring[0].AccountEmail != "soon@x.com" {
		t.Errorf("expiring = %d rows, first=%v", len(expiring), expiring)
	}
}

func TestStorageProviderRepo_Delete(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	ctx := context.Background()

	sp := &model.StorageProvider{
		UserID:               1,
		ProviderType:         "fshare",
		DisplayName:          "to-delete",
		AccountEmail:         "d@x.com",
		CredentialsEncrypted: []byte{0x00},
	}
	_ = repo.Create(ctx, sp)

	if err := repo.Delete(ctx, sp.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, sp.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v; want ErrNotFound after delete", err)
	}

	// Second delete → ErrNotFound
	err = repo.Delete(ctx, sp.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("second delete err = %v; want ErrNotFound", err)
	}
}

func TestStorageProviderRepo_UniqueConstraint(t *testing.T) {
	_, repo := setupStorageProviderTestDB(t)
	ctx := context.Background()

	sp1 := &model.StorageProvider{
		UserID:               1,
		ProviderType:         "fshare",
		DisplayName:          "First",
		AccountEmail:         "u@x.com",
		CredentialsEncrypted: []byte{0x01},
	}
	if err := repo.Create(ctx, sp1); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	sp2 := &model.StorageProvider{
		UserID:               1,
		ProviderType:         "fshare",
		DisplayName:          "Duplicate",
		AccountEmail:         "u@x.com",
		CredentialsEncrypted: []byte{0x02},
	}
	if err := repo.Create(ctx, sp2); err == nil {
		t.Error("expected UNIQUE violation on (user_id, provider_type, account_email)")
	}
}

func TestStorageProviderRepo_ComputeHealth(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		lastErr *string
		expiry  *time.Time
		want    model.HealthStatus
	}{
		{"no expiry", nil, nil, model.HealthUnknown},
		{"expired", nil, timeRef(now.Add(-time.Minute)), model.HealthExpired},
		{"expiring soon (<5min)", nil, timeRef(now.Add(2 * time.Minute)), model.HealthExpiringSoon},
		{"healthy", nil, timeRef(now.Add(20 * time.Minute)), model.HealthHealthy},
		{"error wins", strRef("boom"), timeRef(now.Add(20 * time.Minute)), model.HealthError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.StorageProvider{LastError: tc.lastErr, TokenExpiresAt: tc.expiry}
			got := p.ComputeHealth(now)
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func timeRef(t time.Time) *time.Time { return &t }
func strRef(s string) *string        { return &s }
