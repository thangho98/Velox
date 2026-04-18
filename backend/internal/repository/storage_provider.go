package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/thawng/velox/internal/model"
)

type StorageProviderRepo struct {
	db DBTX
}

func NewStorageProviderRepo(db DBTX) *StorageProviderRepo {
	return &StorageProviderRepo{db: db}
}

// storageProviderColumns is the shared column list for SELECT queries.
const storageProviderColumns = `
	id, user_id, provider_type, display_name, account_email,
	credentials_encrypted, account_type, quota_bytes, used_bytes,
	account_expires_at, token_expires_at, last_refresh_at, last_check_at,
	last_error, created_at`

func scanStorageProvider(row interface{ Scan(...any) error }) (*model.StorageProvider, error) {
	var p model.StorageProvider
	err := row.Scan(
		&p.ID, &p.UserID, &p.ProviderType, &p.DisplayName, &p.AccountEmail,
		&p.CredentialsEncrypted, &p.AccountType, &p.QuotaBytes, &p.UsedBytes,
		&p.AccountExpiresAt, &p.TokenExpiresAt, &p.LastRefreshAt, &p.LastCheckAt,
		&p.LastError, &p.CreatedAt,
	)
	return &p, err
}

// Create inserts a new storage provider. credentials_encrypted must be populated.
// Returns the row with ID set.
func (r *StorageProviderRepo) Create(ctx context.Context, p *model.StorageProvider) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO storage_providers (
			user_id, provider_type, display_name, account_email,
			credentials_encrypted, account_type, quota_bytes, used_bytes,
			account_expires_at, token_expires_at, last_refresh_at, last_check_at, last_error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.UserID, p.ProviderType, p.DisplayName, p.AccountEmail,
		p.CredentialsEncrypted, p.AccountType, p.QuotaBytes, p.UsedBytes,
		p.AccountExpiresAt, p.TokenExpiresAt, p.LastRefreshAt, p.LastCheckAt, p.LastError,
	)
	if err != nil {
		return fmt.Errorf("create storage provider: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	p.ID = id
	// Fetch to populate CreatedAt (DB default).
	row := r.db.QueryRowContext(ctx,
		`SELECT `+storageProviderColumns+` FROM storage_providers WHERE id = ?`, id)
	fetched, err := scanStorageProvider(row)
	if err != nil {
		return fmt.Errorf("refetch after insert: %w", err)
	}
	*p = *fetched
	return nil
}

func (r *StorageProviderRepo) GetByID(ctx context.Context, id int64) (*model.StorageProvider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+storageProviderColumns+` FROM storage_providers WHERE id = ?`, id)
	p, err := scanStorageProvider(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get storage provider %d: %w", id, err)
	}
	return p, nil
}

func (r *StorageProviderRepo) GetByAccount(ctx context.Context, userID int64, providerType, email string) (*model.StorageProvider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+storageProviderColumns+` FROM storage_providers
		 WHERE user_id = ? AND provider_type = ? AND account_email = ?`,
		userID, providerType, email)
	p, err := scanStorageProvider(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get storage provider by account: %w", err)
	}
	return p, nil
}

func (r *StorageProviderRepo) ListByUser(ctx context.Context, userID int64) ([]*model.StorageProvider, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+storageProviderColumns+` FROM storage_providers
		 WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list storage providers: %w", err)
	}
	defer rows.Close()

	var out []*model.StorageProvider
	for rows.Next() {
		p, err := scanStorageProvider(rows)
		if err != nil {
			return nil, fmt.Errorf("scan storage provider: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *StorageProviderRepo) ListAll(ctx context.Context) ([]*model.StorageProvider, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+storageProviderColumns+` FROM storage_providers ORDER BY user_id, created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all storage providers: %w", err)
	}
	defer rows.Close()

	var out []*model.StorageProvider
	for rows.Next() {
		p, err := scanStorageProvider(rows)
		if err != nil {
			return nil, fmt.Errorf("scan storage provider: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListExpiringSoon returns providers whose token_expires_at is before beforeTime.
// Used by the refresh scheduler.
func (r *StorageProviderRepo) ListExpiringSoon(ctx context.Context, beforeTime time.Time) ([]*model.StorageProvider, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+storageProviderColumns+` FROM storage_providers
		 WHERE token_expires_at IS NOT NULL AND token_expires_at < ?
		 ORDER BY token_expires_at ASC`, beforeTime)
	if err != nil {
		return nil, fmt.Errorf("list expiring providers: %w", err)
	}
	defer rows.Close()

	var out []*model.StorageProvider
	for rows.Next() {
		p, err := scanStorageProvider(rows)
		if err != nil {
			return nil, fmt.Errorf("scan storage provider: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Update writes the full mutable state back to the row. Returns ErrNotFound
// when the row does not exist.
func (r *StorageProviderRepo) Update(ctx context.Context, p *model.StorageProvider) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE storage_providers SET
			display_name = ?,
			credentials_encrypted = ?,
			account_type = ?,
			quota_bytes = ?,
			used_bytes = ?,
			account_expires_at = ?,
			token_expires_at = ?,
			last_refresh_at = ?,
			last_check_at = ?,
			last_error = ?
		WHERE id = ?`,
		p.DisplayName, p.CredentialsEncrypted, p.AccountType,
		p.QuotaBytes, p.UsedBytes, p.AccountExpiresAt,
		p.TokenExpiresAt, p.LastRefreshAt, p.LastCheckAt, p.LastError,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("update storage provider %d: %w", p.ID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateCredentials rewrites only the auth-state columns. Used by the refresh
// scheduler to persist a freshly-obtained token without touching display_name.
func (r *StorageProviderRepo) UpdateCredentials(ctx context.Context, id int64, credsEncrypted []byte, tokenExpiresAt time.Time) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE storage_providers SET
			credentials_encrypted = ?,
			token_expires_at = ?,
			last_refresh_at = ?,
			last_error = NULL
		WHERE id = ?`,
		credsEncrypted, tokenExpiresAt, now, id,
	)
	if err != nil {
		return fmt.Errorf("update credentials for provider %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *StorageProviderRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM storage_providers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete storage provider %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
