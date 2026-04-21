package scanner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/nameparser"
	"github.com/thawng/velox/pkg/ophim"
)

// OphimScanner provides functionality to synchronize movies from Ophim
type OphimScanner struct {
	db            *sql.DB
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	seriesRepo    *repository.SeriesRepo
	seasonRepo    *repository.SeasonRepo
	episodeRepo   *repository.EpisodeRepo
	metadataSvc   MetadataMatcher
	client        *ophim.Client
}

// NewOphimScanner creates a new scanner for Ophim
func NewOphimScanner(
	db *sql.DB,
	mediaRepo *repository.MediaRepo,
	mediaFileRepo *repository.MediaFileRepo,
	seriesRepo *repository.SeriesRepo,
	seasonRepo *repository.SeasonRepo,
	episodeRepo *repository.EpisodeRepo,
) *OphimScanner {
	return &OphimScanner{
		db:            db,
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		seriesRepo:    seriesRepo,
		seasonRepo:    seasonRepo,
		episodeRepo:   episodeRepo,
		client:        ophim.New(),
	}
}

// SetMetadataMatcher optional TMDb metadata enrichment service
func (s *OphimScanner) SetMetadataMatcher(m MetadataMatcher) {
	s.metadataSvc = m
}

// SyncRange synchronizes a specified range of pages from Ophim
func (s *OphimScanner) SyncRange(ctx context.Context, libraryID int64, startPage, endPage int) (int, error) {
	added := 0
	for p := startPage; p <= endPage; p++ {
		res, err := s.client.GetRecentMovies(ctx, p)
		if err != nil {
			log.Printf("Ophim sync failed at page %d: %v", p, err)
			return added, fmt.Errorf("ophim sync failed at page %d: %w", p, err)
		}

		for _, item := range res.Data.Items {
			filePath := fmt.Sprintf("ophim://%s", item.Slug)
			// check if exists
			_, err := s.mediaFileRepo.FindCloudFile(ctx, "ophim", item.Slug)
			if err == nil {
				continue // already existed
			}
			if !errors.Is(err, repository.ErrNotFound) {
				continue
			}

			if err := s.persistItem(ctx, libraryID, item, filePath); err != nil {
				log.Printf("Ophim failed to persist %s: %v", item.Slug, err)
			} else {
				added++
				log.Printf("Ophim indexed item: %s", item.Name)
			}
		}
	}
	return added, nil
}

func (s *OphimScanner) persistItem(ctx context.Context, libraryID int64, item ophim.MovieItem, filePath string) error {
	details, err := s.client.GetMovieDetails(ctx, item.Slug)
	if err != nil {
		return fmt.Errorf("failed to get detail for %s: %w", item.Slug, err)
	}

	isSeries := details.Movie.Type == "series" || details.Movie.Type == "hoathinh" || details.Movie.Type == "tvshows"
	if !isSeries && len(details.Episodes) > 0 {
		isSeries = len(details.Episodes[0].ServerData) > 1
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	mediaRepo := s.mediaRepo.WithTx(tx)
	mediaFileRepo := s.mediaFileRepo.WithTx(tx)
	seriesRepo := s.seriesRepo.WithTx(tx)
	seasonRepo := s.seasonRepo.WithTx(tx)
	episodeRepo := s.episodeRepo.WithTx(tx)

	posterPath := model.PosterPath(s.client.GetImageURL(item.PosterURL))
	backdropPath := model.BackdropPath(s.client.GetImageURL(item.ThumbURL))

	var createdMovieMedia *model.Media

	if isSeries {
		// Insert Series
		series := &model.Series{
			LibraryID:    libraryID,
			Title:        item.Name,
			SortTitle:    item.Name,
			Overview:     "Source: Ophim",
			Status:       "Ended",
			PosterPath:   posterPath,
			BackdropPath: backdropPath,
			ThumbPath:    backdropPath,
		}
		if err := seriesRepo.Create(ctx, series); err != nil {
			return err
		}

		// Insert Season 1
		season := &model.Season{
			SeriesID:     series.ID,
			SeasonNumber: 1,
			Title:        "Mùa 1",
			PosterPath:   posterPath,
		}
		if err := seasonRepo.Create(ctx, season); err != nil {
			return err
		}

		if len(details.Episodes) > 0 {
			episodesData := details.Episodes[0].ServerData
			season.EpisodeCount = len(episodesData)
			seasonRepo.Update(ctx, season)

			for idx, epData := range episodesData {
				epName := epData.Name
				if epName == "" {
					epName = fmt.Sprintf("Tập %d", idx+1)
				}

				// Media
				media := &model.Media{
					LibraryID: libraryID,
					MediaType: "episode",
					Title:     fmt.Sprintf("%s - %s", item.Name, epName),
					SortTitle: fmt.Sprintf("%s - %s", item.Name, epName),
					Overview:  "Source: Ophim",
				}
				if err := mediaRepo.Create(ctx, media); err != nil {
					return err
				}

				// Media file
				epFilePath := fmt.Sprintf("ophim://%s/%s", item.Slug, epData.Slug)
				mf := &model.MediaFile{
					MediaID:    media.ID,
					FilePath:   epFilePath,
					IsPrimary:  true,
					FileSize:   0, // External stream
					VideoCodec: "h264",
					AudioCodec: "aac",
					Height:     1080,
					Width:      1920,
				}
				if err := mediaFileRepo.Create(ctx, mf); err != nil {
					return err
				}

				// Episode record
				episodeNum := idx + 1
				// try parse episode string to int if possible
				if epInt, err := strconv.Atoi(epData.Name); err == nil {
					episodeNum = epInt
				}

				episode := &model.Episode{
					SeriesID:      series.ID,
					SeasonID:      season.ID,
					MediaID:       media.ID,
					EpisodeNumber: episodeNum,
					Title:         epName,
					Overview:      "",
					StillPath:     model.StillPath(backdropPath),
				}
				if err := episodeRepo.Create(ctx, episode); err != nil {
					return err
				}
			}
		}
	} else {
		// Movie flow
		createdMovieMedia = &model.Media{
			LibraryID:    libraryID,
			MediaType:    "movie",
			Title:        item.Name,
			SortTitle:    item.Name,
			Overview:     "Source: Ophim",
			PosterPath:   posterPath,
			ThumbPath:    backdropPath,
			BackdropPath: backdropPath,
		}

		if err := mediaRepo.Create(ctx, createdMovieMedia); err != nil {
			return err
		}

		mf := &model.MediaFile{
			MediaID:    createdMovieMedia.ID,
			FilePath:   filePath,
			IsPrimary:  true,
			FileSize:   0, // External stream
			VideoCodec: "h264",
			AudioCodec: "aac",
			Height:     1080,
			Width:      1920,
		}
		if err := mediaFileRepo.Create(ctx, mf); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Non-critical metadata enrichment for Movies outside of transaction
	if !isSeries && s.metadataSvc != nil && createdMovieMedia != nil {
		titleToSearch := item.OriginName
		if titleToSearch == "" {
			titleToSearch = item.Name
		}
		parsed := nameparser.ParsedMedia{
			Title:     titleToSearch,
			Year:      item.Year,
			MediaType: "movie",
		}
		_ = s.metadataSvc.MatchAndPersistMovie(ctx, createdMovieMedia, parsed, filePath, true)
	}

	return nil
}
