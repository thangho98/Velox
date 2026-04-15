package service

import (
	"context"
	"log"

	"github.com/thawng/velox/internal/metadata"
	"github.com/thawng/velox/internal/model"
)

// syncMediaGenres replaces all genres for a media item.
func (s *MetadataService) syncMediaGenres(ctx context.Context, mediaID int64, genres []metadata.GenreInfo) {
	if len(genres) == 0 {
		return
	}

	if err := s.genreRepo.ClearMediaGenres(ctx, mediaID); err != nil {
		log.Printf("Failed to clear media genres: %v", err)
		return
	}

	for _, g := range genres {
		genreID, err := s.ensureGenre(ctx, g)
		if err != nil {
			log.Printf("Failed to ensure genre %q: %v", g.Name, err)
			continue
		}
		if err := s.genreRepo.LinkToMedia(ctx, mediaID, genreID); err != nil {
			log.Printf("Failed to link genre %q to media %d: %v", g.Name, mediaID, err)
		}
	}
}

// syncSeriesGenres replaces all genres for a series.
func (s *MetadataService) syncSeriesGenres(ctx context.Context, seriesID int64, genres []metadata.GenreInfo) {
	if len(genres) == 0 {
		return
	}

	if err := s.genreRepo.ClearSeriesGenres(ctx, seriesID); err != nil {
		log.Printf("Failed to clear series genres: %v", err)
		return
	}

	for _, g := range genres {
		genreID, err := s.ensureGenre(ctx, g)
		if err != nil {
			log.Printf("Failed to ensure genre %q: %v", g.Name, err)
			continue
		}
		if err := s.genreRepo.LinkToSeries(ctx, seriesID, genreID); err != nil {
			log.Printf("Failed to link genre %q to series %d: %v", g.Name, seriesID, err)
		}
	}
}

// ensureGenre gets or creates a genre, returning its ID.
func (s *MetadataService) ensureGenre(ctx context.Context, g metadata.GenreInfo) (int64, error) {
	// Try by TMDb ID first
	if g.ID > 0 {
		existing, err := s.genreRepo.GetByTmdbID(ctx, int64(g.ID))
		if err == nil {
			return existing.ID, nil
		}
	}

	// Try by name
	existing, err := s.genreRepo.GetByName(ctx, g.Name)
	if err == nil {
		return existing.ID, nil
	}

	// Create new genre
	var tmdbID *int64
	if g.ID > 0 {
		id := int64(g.ID)
		tmdbID = &id
	}
	genre := &model.Genre{
		Name:   g.Name,
		TmdbID: tmdbID,
	}
	if err := s.genreRepo.Create(ctx, genre); err != nil {
		return 0, err
	}
	return genre.ID, nil
}

// syncMediaGenresByName replaces all genres for a media item using genre names (no TMDb ID).
func (s *MetadataService) syncMediaGenresByName(ctx context.Context, mediaID int64, names []string) error {
	if err := s.genreRepo.ClearMediaGenres(ctx, mediaID); err != nil {
		return err
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		genreID, err := s.ensureGenreByName(ctx, name)
		if err != nil {
			log.Printf("Failed to ensure genre %q: %v", name, err)
			continue
		}
		if err := s.genreRepo.LinkToMedia(ctx, mediaID, genreID); err != nil {
			log.Printf("Failed to link genre %q to media %d: %v", name, mediaID, err)
		}
	}
	return nil
}

// syncSeriesGenresByName replaces all genres for a series using genre names.
func (s *MetadataService) syncSeriesGenresByName(ctx context.Context, seriesID int64, names []string) error {
	if err := s.genreRepo.ClearSeriesGenres(ctx, seriesID); err != nil {
		return err
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		genreID, err := s.ensureGenreByName(ctx, name)
		if err != nil {
			log.Printf("Failed to ensure genre %q: %v", name, err)
			continue
		}
		if err := s.genreRepo.LinkToSeries(ctx, seriesID, genreID); err != nil {
			log.Printf("Failed to link genre %q to series %d: %v", name, seriesID, err)
		}
	}
	return nil
}

// ensureGenreByName gets or creates a genre by name (no TMDb ID).
func (s *MetadataService) ensureGenreByName(ctx context.Context, name string) (int64, error) {
	existing, err := s.genreRepo.GetByName(ctx, name)
	if err == nil {
		return existing.ID, nil
	}
	genre := &model.Genre{Name: name}
	if err := s.genreRepo.Create(ctx, genre); err != nil {
		return 0, err
	}
	return genre.ID, nil
}

// syncMediaCredits replaces all credits for a media item.
func (s *MetadataService) syncMediaCredits(ctx context.Context, mediaID int64, cast []metadata.CastInfo, crew []metadata.CrewInfo) {
	if len(cast) == 0 && len(crew) == 0 {
		return
	}

	if err := s.personRepo.ClearMediaCredits(ctx, mediaID); err != nil {
		log.Printf("Failed to clear media credits: %v", err)
		return
	}

	// Top 20 cast
	limit := 20
	if len(cast) < limit {
		limit = len(cast)
	}
	for i := 0; i < limit; i++ {
		c := cast[i]
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		credit := &model.Credit{
			MediaID:      &mediaID,
			PersonID:     personID,
			Character:    c.Character,
			Role:         "cast",
			DisplayOrder: c.Order,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add cast credit: %v", err)
		}
	}

	// Key crew (director, writer)
	for _, c := range crew {
		if c.Job != "Director" && c.Job != "Writer" && c.Job != "Screenplay" {
			continue
		}
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		role := "director"
		if c.Job == "Writer" || c.Job == "Screenplay" {
			role = "writer"
		}
		credit := &model.Credit{
			MediaID:  &mediaID,
			PersonID: personID,
			Role:     role,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add crew credit: %v", err)
		}
	}
}

// syncSeriesCredits replaces all credits for a series.
func (s *MetadataService) syncSeriesCredits(ctx context.Context, seriesID int64, cast []metadata.CastInfo, crew []metadata.CrewInfo) {
	if len(cast) == 0 && len(crew) == 0 {
		return
	}

	if err := s.personRepo.ClearSeriesCredits(ctx, seriesID); err != nil {
		log.Printf("Failed to clear series credits: %v", err)
		return
	}

	limit := 20
	if len(cast) < limit {
		limit = len(cast)
	}
	for i := 0; i < limit; i++ {
		c := cast[i]
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		credit := &model.Credit{
			SeriesID:     &seriesID,
			PersonID:     personID,
			Character:    c.Character,
			Role:         "cast",
			DisplayOrder: c.Order,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add series cast credit: %v", err)
		}
	}

	for _, c := range crew {
		if c.Job != "Director" && c.Job != "Writer" && c.Job != "Screenplay" {
			continue
		}
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		role := "director"
		if c.Job == "Writer" || c.Job == "Screenplay" {
			role = "writer"
		}
		credit := &model.Credit{
			SeriesID: &seriesID,
			PersonID: personID,
			Role:     role,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add series crew credit: %v", err)
		}
	}
}

// ensurePerson gets or creates a person, returning their local ID.
func (s *MetadataService) ensurePerson(ctx context.Context, tmdbPersonID int, name, profilePath string) (int64, error) {
	if tmdbPersonID > 0 {
		existing, err := s.personRepo.GetByTmdbID(ctx, int64(tmdbPersonID))
		if err == nil {
			return existing.ID, nil
		}
	}

	var tmdbID *int64
	if tmdbPersonID > 0 {
		id := int64(tmdbPersonID)
		tmdbID = &id
	}
	person := &model.Person{
		Name:        name,
		TmdbID:      tmdbID,
		ProfilePath: model.ProfilePath(profilePath),
	}
	if err := s.personRepo.Create(ctx, person); err != nil {
		return 0, err
	}
	return person.ID, nil
}

// syncMediaCreditsByInput replaces all credits for a media item using CreditInput.
func (s *MetadataService) syncMediaCreditsByInput(ctx context.Context, mediaID int64, credits []model.CreditInput) error {
	if err := s.personRepo.ClearMediaCredits(ctx, mediaID); err != nil {
		return err
	}
	for _, c := range credits {
		if c.PersonName == "" {
			continue
		}
		personID, err := s.ensurePersonByName(ctx, c.PersonName)
		if err != nil {
			log.Printf("Failed to ensure person %q: %v", c.PersonName, err)
			continue
		}
		credit := &model.Credit{
			MediaID:      &mediaID,
			PersonID:     personID,
			Character:    c.Character,
			Role:         c.Role,
			DisplayOrder: c.Order,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add credit for %q: %v", c.PersonName, err)
		}
	}
	return nil
}

// syncSeriesCreditsByInput replaces all credits for a series using CreditInput.
func (s *MetadataService) syncSeriesCreditsByInput(ctx context.Context, seriesID int64, credits []model.CreditInput) error {
	if err := s.personRepo.ClearSeriesCredits(ctx, seriesID); err != nil {
		return err
	}
	for _, c := range credits {
		if c.PersonName == "" {
			continue
		}
		personID, err := s.ensurePersonByName(ctx, c.PersonName)
		if err != nil {
			log.Printf("Failed to ensure person %q: %v", c.PersonName, err)
			continue
		}
		credit := &model.Credit{
			SeriesID:     &seriesID,
			PersonID:     personID,
			Character:    c.Character,
			Role:         c.Role,
			DisplayOrder: c.Order,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add credit for %q: %v", c.PersonName, err)
		}
	}
	return nil
}

// ensurePersonByName gets or creates a person by name (no TMDb ID).
func (s *MetadataService) ensurePersonByName(ctx context.Context, name string) (int64, error) {
	existing, err := s.personRepo.GetByName(ctx, name)
	if err == nil {
		return existing.ID, nil
	}
	person := &model.Person{Name: name}
	if err := s.personRepo.Create(ctx, person); err != nil {
		return 0, err
	}
	return person.ID, nil
}
