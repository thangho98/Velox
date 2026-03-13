package nameparser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     ParsedMedia
	}{
		{
			name:     "Standard movie with year and quality",
			filename: "The.Matrix.1999.1080p.BluRay.x264.mkv",
			want: ParsedMedia{
				Title:     "The Matrix",
				Year:      1999,
				Season:    0,
				Episode:   0,
				MediaType: "movie",
				Quality:   "1080p",
				Codec:     "X264",
			},
		},
		{
			name:     "Movie with spaces",
			filename: "Inception 2010 720p BluRay.mp4",
			want: ParsedMedia{
				Title:     "Inception",
				Year:      2010,
				Season:    0,
				Episode:   0,
				MediaType: "movie",
				Quality:   "720p",
			},
		},
		{
			name:     "Movie with 4K",
			filename: "Dune.2021.4K.HDR.HEVC.mkv",
			want: ParsedMedia{
				Title:     "Dune",
				Year:      2021,
				Season:    0,
				Episode:   0,
				MediaType: "movie",
				Quality:   "4K",
				Codec:     "HEVC",
			},
		},

		// TV Series - Standard format
		{
			name:     "Series S01E01 format",
			filename: "Breaking.Bad.S01E01.Pilot.720p.mkv",
			want: ParsedMedia{
				Title:        "Breaking Bad",
				EpisodeTitle: "Pilot",
				Season:       1,
				Episode:      1,
				MediaType:    "episode",
				Quality:      "720p",
			},
		},
		{
			name:     "Series s01e01 lowercase",
			filename: "game.of.thrones.s01e01.1080p.mkv",
			want: ParsedMedia{
				Title:     "game of thrones",
				Season:    1,
				Episode:   1,
				MediaType: "episode",
				Quality:   "1080p",
			},
		},
		{
			name:     "Series with year and SxxExx",
			filename: "Doctor.Who.2005.S12E01.720p.mkv",
			want: ParsedMedia{
				Title:     "Doctor Who",
				Year:      2005,
				Season:    12,
				Episode:   1,
				MediaType: "episode",
				Quality:   "720p",
			},
		},

		// TV Series - Alternative format
		{
			name:     "Series 1x01 format",
			filename: "Friends.1x01.The.One.Where.It.All.Began.avi",
			want: ParsedMedia{
				Title:        "Friends",
				EpisodeTitle: "The One Where It All Began",
				Season:       1,
				Episode:      1,
				MediaType:    "episode",
			},
		},
		{
			name:     "Series 10x05 double digit season",
			filename: "The.Simpsons.10x05.480p.mkv",
			want: ParsedMedia{
				Title:     "The Simpsons",
				Season:    10,
				Episode:   5,
				MediaType: "episode",
				Quality:   "480p",
			},
		},

		// Multi-episode
		{
			name:     "Multi-episode S01E01E02",
			filename: "Show.S01E01E02.Title.mkv",
			want: ParsedMedia{
				Title:        "Show",
				EpisodeTitle: "Title",
				Season:       1,
				Episode:      1,
				EndEpisode:   2,
				MediaType:    "episode",
			},
		},
		{
			name:     "Multi-episode with dash",
			filename: "Anime.S01E01-E03.Batch.mkv",
			want: ParsedMedia{
				Title:        "Anime",
				EpisodeTitle: "E03 Batch",
				Season:       1,
				Episode:      1,
				EndEpisode:   3,
				MediaType:    "episode",
			},
		},

		// Daily shows / Date format
		{
			name:     "Daily show with date",
			filename: "The.Daily.Show.2024.03.15.1080p.mkv",
			want: ParsedMedia{
				Title:     "The Daily Show",
				Year:      2024,
				Season:    0,
				Episode:   0,
				MediaType: "movie",
				Quality:   "1080p",
			},
		},

		// Release group
		{
			name:     "Movie with release group brackets",
			filename: "Movie.2020.1080p.[YTS.MX].mp4",
			want: ParsedMedia{
				Title:        "Movie",
				Year:         2020,
				MediaType:    "movie",
				Quality:      "1080p",
				ReleaseGroup: "YTS.MX",
			},
		},
		{
			name:     "Movie with release group parens",
			filename: "Movie.2020.1080p.(RARBG).mkv",
			want: ParsedMedia{
				Title:        "Movie",
				Year:         2020,
				MediaType:    "movie",
				Quality:      "1080p",
				ReleaseGroup: "RARBG",
			},
		},

		// Edge cases
		{
			name:     "No year no quality",
			filename: "Some.Random.Movie.mkv",
			want: ParsedMedia{
				Title:     "Some Random Movie",
				Year:      0,
				MediaType: "movie",
			},
		},
		{
			name:     "Extra dots in title",
			filename: "S.W.A.T.2017.S01E01.720p.mkv",
			want: ParsedMedia{
				Title:     "S W A T",
				Year:      2017,
				Season:    1,
				Episode:   1,
				MediaType: "episode",
				Quality:   "720p",
			},
		},
		{
			name:     "Unicode title",
			filename: "Parásitos.2019.1080p.mkv",
			want: ParsedMedia{
				Title:     "Parásitos",
				Year:      2019,
				MediaType: "movie",
				Quality:   "1080p",
			},
		},
		{
			name:     "HEVC codec detection",
			filename: "Movie.2020.2160p.HEVC.x265.mkv",
			want: ParsedMedia{
				Title:     "Movie",
				Year:      2020,
				MediaType: "movie",
				Quality:   "2160p",
				Codec:     "HEVC",
			},
		},
		{
			name:     "VP9 codec",
			filename: "Movie.2021.1080p.VP9.webm",
			want: ParsedMedia{
				Title:     "Movie",
				Year:      2021,
				MediaType: "movie",
				Quality:   "1080p",
				Codec:     "VP9",
			},
		},
		{
			name:     "Anime triple episode",
			filename: "Anime.S01E01E02E03.1080p.mkv",
			want: ParsedMedia{
				Title:        "Anime",
				EpisodeTitle: "E03",
				Season:       1,
				Episode:      1,
				EndEpisode:   2,
				MediaType:    "episode",
				Quality:      "1080p",
			},
		},
		// Real-world test: Malcolm in the Middle format
		{
			name:     "Malcolm in the Middle AMZN WEB-DL",
			filename: "Malcolm in the Middle (2000) - S01E01 - Pilot (1080p AMZN WEB-DL x265 Silence).mkv",
			want: ParsedMedia{
				Title:        "Malcolm in the Middle",
				EpisodeTitle: "Pilot",
				Year:         2000,
				Season:       1,
				Episode:      1,
				MediaType:    "episode",
				Quality:      "1080p",
				Codec:        "X265",
				ReleaseGroup: "Silence",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.filename)

			if got.Title != tt.want.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.want.Title)
			}
			if got.EpisodeTitle != tt.want.EpisodeTitle {
				t.Errorf("EpisodeTitle = %q, want %q", got.EpisodeTitle, tt.want.EpisodeTitle)
			}
			if got.Year != tt.want.Year {
				t.Errorf("Year = %d, want %d", got.Year, tt.want.Year)
			}
			if got.Season != tt.want.Season {
				t.Errorf("Season = %d, want %d", got.Season, tt.want.Season)
			}
			if got.Episode != tt.want.Episode {
				t.Errorf("Episode = %d, want %d", got.Episode, tt.want.Episode)
			}
			if got.MediaType != tt.want.MediaType {
				t.Errorf("MediaType = %q, want %q", got.MediaType, tt.want.MediaType)
			}
			if got.Quality != tt.want.Quality {
				t.Errorf("Quality = %q, want %q", got.Quality, tt.want.Quality)
			}
			if got.Codec != tt.want.Codec {
				t.Errorf("Codec = %q, want %q", got.Codec, tt.want.Codec)
			}
			if got.ReleaseGroup != tt.want.ReleaseGroup {
				t.Errorf("ReleaseGroup = %q, want %q", got.ReleaseGroup, tt.want.ReleaseGroup)
			}
		})
	}
}

func TestParseWithParents(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantType   string
		wantSeason int
	}{
		{
			name:       "Series from parent folder",
			path:       "/media/TV Shows/Game of Thrones/Season 1/episode.mkv",
			wantType:   "episode",
			wantSeason: 1,
		},
		{
			name:       "Series from parent folder Season 05",
			path:       "/media/TV/Breaking Bad/Season 05/Episode.mkv",
			wantType:   "episode",
			wantSeason: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseWithParents(tt.path)
			if got.MediaType != tt.wantType {
				t.Errorf("MediaType = %q, want %q", got.MediaType, tt.wantType)
			}
			if got.Season != tt.wantSeason {
				t.Errorf("Season = %d, want %d", got.Season, tt.wantSeason)
			}
		})
	}
}

func TestIsMultiEpisode(t *testing.T) {
	tests := []struct {
		name string
		p    ParsedMedia
		want bool
	}{
		{
			name: "Single episode",
			p:    ParsedMedia{Episode: 1, EndEpisode: 0},
			want: false,
		},
		{
			name: "Multi episode",
			p:    ParsedMedia{Episode: 1, EndEpisode: 3},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsMultiEpisode(); got != tt.want {
				t.Errorf("IsMultiEpisode() = %v, want %v", got, tt.want)
			}
		})
	}
}
