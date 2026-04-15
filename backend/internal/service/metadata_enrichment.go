package service

import (
	"context"
	"log"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/pkg/fanart"
	"github.com/thawng/velox/pkg/thetvdb"
	"github.com/thawng/velox/pkg/tvmaze"
)

// enrichOMDbRatings fetches IMDb/RT/Metacritic ratings from OMDb and updates media.
// Non-fatal: logs errors and continues.
func (s *MetadataService) enrichOMDbRatings(ctx context.Context, media *model.Media) {
	if s.omdbClient == nil || media.ImdbID == nil || *media.ImdbID == "" {
		return
	}

	result, err := s.omdbClient.GetByIMDbID(ctx, *media.ImdbID)
	if err != nil {
		log.Printf("OMDb lookup failed for %s: %v", *media.ImdbID, err)
		return
	}

	media.IMDbRating = result.IMDbRatingFloat()
	media.RTScore = result.RottenTomatoesScore()
	media.MetacriticScore = result.MetascoreInt()

	if err := s.mediaRepo.UpdateOMDbRatings(ctx, media.ID, media.IMDbRating, media.RTScore, media.MetacriticScore); err != nil {
		log.Printf("Failed to save OMDb ratings for media %d: %v", media.ID, err)
	}
}

// enrichTVDBSeries fetches series info from TheTVDB to fill in missing data
// (e.g. IMDb ID from remote IDs). Non-fatal: logs errors and continues.
func (s *MetadataService) enrichTVDBSeries(ctx context.Context, series *model.Series) {
	if s.tvdbClient == nil || series.TvdbID == nil || *series.TvdbID == 0 {
		return
	}

	tvdbSeries, err := s.tvdbClient.GetSeriesExtended(ctx, int(*series.TvdbID), true)
	if err != nil {
		log.Printf("TheTVDB lookup failed for series %d (tvdb:%d): %v", series.ID, *series.TvdbID, err)
		return
	}

	changed := false

	// Fill in IMDb ID if missing
	if (series.ImdbID == nil || *series.ImdbID == "") && len(tvdbSeries.RemoteIDs) > 0 {
		imdbID := thetvdb.IMDbID(tvdbSeries.RemoteIDs)
		if imdbID != "" {
			series.ImdbID = &imdbID
			changed = true
		}
	}

	// Fill in status if missing
	if series.Status == "" && tvdbSeries.Status != nil {
		series.Status = tvdbSeries.Status.Name
		changed = true
	}

	if changed {
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			log.Printf("Failed to update series %d with TVDB data: %v", series.ID, err)
		}
	}
}

// enrichFanartMovie fetches artwork from fanart.tv for a movie (by TMDb ID).
// Non-fatal: logs errors and continues.
func (s *MetadataService) enrichFanartMovie(ctx context.Context, media *model.Media) {
	if s.fanartClient == nil || media.TmdbID == nil || *media.TmdbID == 0 {
		return
	}

	images, err := s.fanartClient.GetMovieImages(ctx, int(*media.TmdbID))
	if err != nil {
		if err != fanart.ErrNotFound {
			log.Printf("fanart.tv movie lookup failed for tmdb:%d: %v", *media.TmdbID, err)
		}
		return
	}

	changed := false

	if media.LogoPath == "" {
		if logo := fanart.BestImage(images.HDClearLogo); logo != "" {
			media.LogoPath = model.LogoPath(logo)
			changed = true
		} else if logo := fanart.BestImage(images.MovieLogo); logo != "" {
			media.LogoPath = model.LogoPath(logo)
			changed = true
		}
	}

	if media.ThumbPath == "" {
		if thumb := fanart.BestImage(images.MovieThumb); thumb != "" {
			media.ThumbPath = model.BackdropPath(thumb)
			changed = true
		}
	}

	if changed {
		if err := s.mediaRepo.Update(ctx, media); err != nil {
			log.Printf("Failed to save fanart.tv artwork for media %d: %v", media.ID, err)
		} else {
			s.enqueueImageForCompute(string(media.LogoPath))
			s.enqueueImageForCompute(string(media.ThumbPath))
		}
	}
}

// enrichFanartShow fetches artwork from fanart.tv for a TV show (by TVDB ID).
// Non-fatal: logs errors and continues.
func (s *MetadataService) enrichFanartShow(ctx context.Context, series *model.Series) {
	if s.fanartClient == nil || series.TvdbID == nil || *series.TvdbID == 0 {
		return
	}

	images, err := s.fanartClient.GetShowImages(ctx, int(*series.TvdbID))
	if err != nil {
		if err != fanart.ErrNotFound {
			log.Printf("fanart.tv show lookup failed for tvdb:%d: %v", *series.TvdbID, err)
		}
		return
	}

	changed := false

	if series.LogoPath == "" {
		if logo := fanart.BestImage(images.HDClearLogo); logo != "" {
			series.LogoPath = model.LogoPath(logo)
			changed = true
		} else if logo := fanart.BestImage(images.ClearLogo); logo != "" {
			series.LogoPath = model.LogoPath(logo)
			changed = true
		}
	}

	if series.ThumbPath == "" {
		if thumb := fanart.BestImage(images.TVThumb); thumb != "" {
			series.ThumbPath = model.BackdropPath(thumb)
			changed = true
		}
	}

	if changed {
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			log.Printf("Failed to save fanart.tv artwork for series %d: %v", series.ID, err)
		} else {
			s.enqueueImageForCompute(string(series.LogoPath))
			s.enqueueImageForCompute(string(series.ThumbPath))
		}
	}
}

// enrichTVmazeSeries fetches series info from TVmaze to fill in network and missing IDs.
// Looks up by TVDB ID first, then IMDb ID. Non-fatal: logs errors and continues.
func (s *MetadataService) enrichTVmazeSeries(ctx context.Context, series *model.Series) {
	if s.tvmazeClient == nil {
		return
	}

	var show *tvmaze.Show
	var err error

	// Try TVDB ID first
	if series.TvdbID != nil && *series.TvdbID > 0 {
		show, err = s.tvmazeClient.LookupByTVDB(ctx, int(*series.TvdbID))
		if err != nil && err != tvmaze.ErrNotFound {
			log.Printf("TVmaze TVDB lookup failed for series %d (tvdb:%d): %v", series.ID, *series.TvdbID, err)
		}
	}

	// Fallback to IMDb ID
	if show == nil && series.ImdbID != nil && *series.ImdbID != "" {
		show, err = s.tvmazeClient.LookupByIMDb(ctx, *series.ImdbID)
		if err != nil && err != tvmaze.ErrNotFound {
			log.Printf("TVmaze IMDb lookup failed for series %d (%s): %v", series.ID, *series.ImdbID, err)
		}
	}

	if show == nil {
		return
	}

	changed := false

	// Fill in network if missing
	if series.Network == "" {
		network := show.NetworkName()
		if network != "" {
			series.Network = network
			changed = true
		}
	}

	// Fill in status if missing
	if series.Status == "" && show.Status != "" {
		series.Status = show.Status
		changed = true
	}

	// Fill in TVDB ID if missing
	if (series.TvdbID == nil || *series.TvdbID == 0) && show.Externals.TheTVDB != nil {
		tvdbID := int64(*show.Externals.TheTVDB)
		series.TvdbID = &tvdbID
		changed = true
	}

	// Fill in IMDb ID if missing
	if (series.ImdbID == nil || *series.ImdbID == "") && show.Externals.IMDb != "" {
		series.ImdbID = &show.Externals.IMDb
		changed = true
	}

	if changed {
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			log.Printf("Failed to update series %d with TVmaze data: %v", series.ID, err)
		}
	}
}
