package anilist

import (
	"fmt"
	"strings"
)

// MediaPage represents a paginated AniList media response.
type MediaPage struct {
	PageInfo PageInfo `json:"pageInfo"`
	Media    []Media  `json:"media"`
}

// PageInfo contains pagination metadata.
type PageInfo struct {
	Total       int  `json:"total"`
	PerPage     int  `json:"perPage"`
	CurrentPage int  `json:"currentPage"`
	LastPage    int  `json:"lastPage"`
	HasNextPage bool `json:"hasNextPage"`
}

// Media contains the AniList media fields needed by the anime metadata flow.
type Media struct {
	ID                int                `json:"id"`
	IDMal             *int               `json:"idMal,omitempty"`
	Type              string             `json:"type"`
	Format            string             `json:"format"`
	Status            string             `json:"status"`
	Episodes          *int               `json:"episodes,omitempty"`
	Season            string             `json:"season"`
	SeasonYear        *int               `json:"seasonYear,omitempty"`
	Description       string             `json:"description"`
	SiteURL           string             `json:"siteUrl"`
	Title             MediaTitle         `json:"title"`
	CoverImage        CoverImage         `json:"coverImage"`
	BannerImage       string             `json:"bannerImage"`
	StartDate         FuzzyDate          `json:"startDate"`
	EndDate           FuzzyDate          `json:"endDate"`
	Studios           StudioConnection   `json:"studios"`
	StreamingEpisodes []StreamingEpisode `json:"streamingEpisodes"`
	NextAiringEpisode *NextAiringEpisode `json:"nextAiringEpisode,omitempty"`
}

// MediaTitle contains AniList title variants.
type MediaTitle struct {
	Romaji        string `json:"romaji"`
	English       string `json:"english"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

// Preferred returns the best available title for display and matching.
func (t MediaTitle) Preferred() string {
	for _, candidate := range []string{t.UserPreferred, t.English, t.Romaji, t.Native} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}

// CoverImage contains AniList artwork URLs.
type CoverImage struct {
	ExtraLarge string `json:"extraLarge"`
	Large      string `json:"large"`
	Medium     string `json:"medium"`
	Color      string `json:"color"`
}

// StudioConnection contains the studios attached to a media item.
type StudioConnection struct {
	Nodes []Studio `json:"nodes"`
}

// PrimaryName returns the best studio name for metadata backfill.
func (c StudioConnection) PrimaryName() string {
	for _, studio := range c.Nodes {
		if studio.IsAnimationStudio && strings.TrimSpace(studio.Name) != "" {
			return studio.Name
		}
	}
	for _, studio := range c.Nodes {
		if strings.TrimSpace(studio.Name) != "" {
			return studio.Name
		}
	}
	return ""
}

// Studio represents an AniList studio node.
type Studio struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	IsAnimationStudio bool   `json:"isAnimationStudio"`
}

// StreamingEpisode contains sparse episode metadata AniList exposes for legal streams.
type StreamingEpisode struct {
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	URL       string `json:"url"`
	Site      string `json:"site"`
}

// NextAiringEpisode contains schedule info for upcoming episodes.
type NextAiringEpisode struct {
	AiringAt        int64 `json:"airingAt"`
	TimeUntilAiring int   `json:"timeUntilAiring"`
	Episode         int   `json:"episode"`
}

// FuzzyDate is AniList's partial date type.
type FuzzyDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// String converts the fuzzy date into a best-effort YYYY-MM-DD string.
func (d FuzzyDate) String() string {
	if d.Year <= 0 {
		return ""
	}
	month := d.Month
	day := d.Day
	if month <= 0 {
		month = 1
	}
	if day <= 0 {
		day = 1
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, month, day)
}

// PreferredTitle returns the best available title for the media.
func (m Media) PreferredTitle() string {
	return m.Title.Preferred()
}

// PrimaryStudio returns the best available studio name.
func (m Media) PrimaryStudio() string {
	return m.Studios.PrimaryName()
}
