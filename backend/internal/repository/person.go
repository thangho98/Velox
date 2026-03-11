package repository

import (
	"context"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// PersonRepo handles people database operations
type PersonRepo struct {
	db DBTX
}

func NewPersonRepo(db DBTX) *PersonRepo {
	return &PersonRepo{db: db}
}

// Create inserts a new person
func (r *PersonRepo) Create(ctx context.Context, p *model.Person) error {
	query := `INSERT INTO people (name, tmdb_id, profile_path) VALUES (?, ?, ?) RETURNING id`
	row := r.db.QueryRowContext(ctx, query, p.Name, p.TmdbID, p.ProfilePath)
	return row.Scan(&p.ID)
}

// GetByID retrieves a person by ID
func (r *PersonRepo) GetByID(ctx context.Context, id int64) (*model.Person, error) {
	var p model.Person
	err := r.db.QueryRowContext(ctx, "SELECT id, name, tmdb_id, profile_path FROM people WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.TmdbID, &p.ProfilePath)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByName retrieves a person by name
func (r *PersonRepo) GetByName(ctx context.Context, name string) (*model.Person, error) {
	var p model.Person
	err := r.db.QueryRowContext(ctx, "SELECT id, name, tmdb_id, profile_path FROM people WHERE name = ?", name).
		Scan(&p.ID, &p.Name, &p.TmdbID, &p.ProfilePath)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByTmdbID retrieves a person by TMDb ID
func (r *PersonRepo) GetByTmdbID(ctx context.Context, tmdbID int64) (*model.Person, error) {
	var p model.Person
	err := r.db.QueryRowContext(ctx, "SELECT id, name, tmdb_id, profile_path FROM people WHERE tmdb_id = ?", tmdbID).
		Scan(&p.ID, &p.Name, &p.TmdbID, &p.ProfilePath)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update updates a person
func (r *PersonRepo) Update(ctx context.Context, p *model.Person) error {
	_, err := r.db.ExecContext(ctx, "UPDATE people SET name = ?, tmdb_id = ?, profile_path = ? WHERE id = ?",
		p.Name, p.TmdbID, p.ProfilePath, p.ID)
	return err
}

// Delete removes a person
func (r *PersonRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM people WHERE id = ?", id)
	return err
}

// Search searches people by name
func (r *PersonRepo) Search(ctx context.Context, query string, limit int) ([]model.Person, error) {
	q := "SELECT id, name, tmdb_id, profile_path FROM people WHERE name LIKE ? ORDER BY name LIMIT ?"
	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, q, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("searching people: %w", err)
	}
	defer rows.Close()

	var items []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.TmdbID, &p.ProfilePath); err != nil {
			return nil, fmt.Errorf("scanning person: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// AddCredit adds a credit (role) for a person in media or series
func (r *PersonRepo) AddCredit(ctx context.Context, c *model.Credit) error {
	query := `INSERT INTO credits
		(media_id, series_id, person_id, character, role, display_order)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`

	row := r.db.QueryRowContext(ctx, query,
		c.MediaID, c.SeriesID, c.PersonID, c.Character, c.Role, c.DisplayOrder)
	return row.Scan(&c.ID)
}

// UpdateCredit updates a credit
func (r *PersonRepo) UpdateCredit(ctx context.Context, c *model.Credit) error {
	_, err := r.db.ExecContext(ctx, `UPDATE credits SET
		media_id = ?, series_id = ?, person_id = ?, character = ?, role = ?, display_order = ?
		WHERE id = ?`,
		c.MediaID, c.SeriesID, c.PersonID, c.Character, c.Role, c.DisplayOrder, c.ID)
	return err
}

// RemoveCredit removes a credit by ID
func (r *PersonRepo) RemoveCredit(ctx context.Context, creditID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM credits WHERE id = ?", creditID)
	return err
}

// ListCreditsByMedia retrieves all credits for a media item with person details
func (r *PersonRepo) ListCreditsByMedia(ctx context.Context, mediaID int64) ([]model.CreditWithPerson, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT c.id, c.media_id, c.series_id, c.person_id,
		c.character, c.role, c.display_order, p.name, p.tmdb_id, p.profile_path
		FROM credits c
		JOIN people p ON p.id = c.person_id
		WHERE c.media_id = ?
		ORDER BY c.display_order, p.name`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("listing credits by media: %w", err)
	}
	defer rows.Close()

	var items []model.CreditWithPerson
	for rows.Next() {
		var cp model.CreditWithPerson
		if err := rows.Scan(&cp.Credit.ID, &cp.Credit.MediaID, &cp.Credit.SeriesID, &cp.Credit.PersonID,
			&cp.Credit.Character, &cp.Credit.Role, &cp.Credit.DisplayOrder,
			&cp.Person.Name, &cp.Person.TmdbID, &cp.Person.ProfilePath); err != nil {
			return nil, fmt.Errorf("scanning credit: %w", err)
		}
		cp.Person.ID = cp.Credit.PersonID
		items = append(items, cp)
	}
	return items, rows.Err()
}

// ListCreditsBySeries retrieves all credits for a series with person details
func (r *PersonRepo) ListCreditsBySeries(ctx context.Context, seriesID int64) ([]model.CreditWithPerson, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT c.id, c.media_id, c.series_id, c.person_id,
		c.character, c.role, c.display_order, p.name, p.tmdb_id, p.profile_path
		FROM credits c
		JOIN people p ON p.id = c.person_id
		WHERE c.series_id = ?
		ORDER BY c.display_order, p.name`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("listing credits by series: %w", err)
	}
	defer rows.Close()

	var items []model.CreditWithPerson
	for rows.Next() {
		var cp model.CreditWithPerson
		if err := rows.Scan(&cp.Credit.ID, &cp.Credit.MediaID, &cp.Credit.SeriesID, &cp.Credit.PersonID,
			&cp.Credit.Character, &cp.Credit.Role, &cp.Credit.DisplayOrder,
			&cp.Person.Name, &cp.Person.TmdbID, &cp.Person.ProfilePath); err != nil {
			return nil, fmt.Errorf("scanning credit: %w", err)
		}
		cp.Person.ID = cp.Credit.PersonID
		items = append(items, cp)
	}
	return items, rows.Err()
}

// ListCreditsByPerson retrieves all credits for a person
func (r *PersonRepo) ListCreditsByPerson(ctx context.Context, personID int64) ([]model.Credit, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_id, series_id, person_id,
		character, role, display_order
		FROM credits WHERE person_id = ?
		ORDER BY display_order`, personID)
	if err != nil {
		return nil, fmt.Errorf("listing credits by person: %w", err)
	}
	defer rows.Close()

	var items []model.Credit
	for rows.Next() {
		var c model.Credit
		if err := rows.Scan(&c.ID, &c.MediaID, &c.SeriesID, &c.PersonID,
			&c.Character, &c.Role, &c.DisplayOrder); err != nil {
			return nil, fmt.Errorf("scanning credit: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// ClearMediaCredits removes all credits for a media item
func (r *PersonRepo) ClearMediaCredits(ctx context.Context, mediaID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM credits WHERE media_id = ?", mediaID)
	return err
}

// ClearSeriesCredits removes all credits for a series
func (r *PersonRepo) ClearSeriesCredits(ctx context.Context, seriesID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM credits WHERE series_id = ?", seriesID)
	return err
}
