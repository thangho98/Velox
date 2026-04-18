package anilist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSearchAnime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("content-type = %q, want json", got)
		}

		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "streamingEpisodes") {
			t.Fatalf("query missing streamingEpisodes field")
		}
		if req.Variables["search"] != "Frieren" {
			t.Fatalf("search variable = %v, want Frieren", req.Variables["search"])
		}

		_, _ = w.Write([]byte(`{
			"data": {
				"Page": {
					"pageInfo": {
						"total": 1,
						"perPage": 10,
						"currentPage": 1,
						"lastPage": 1,
						"hasNextPage": false
					},
					"media": [{
						"id": 52991,
						"type": "ANIME",
						"format": "TV",
						"status": "FINISHED",
						"episodes": 28,
						"season": "FALL",
						"seasonYear": 2023,
						"description": "An elf mage remembers her past.",
						"siteUrl": "https://anilist.co/anime/52991",
						"title": {
							"romaji": "Sousou no Frieren",
							"english": "Frieren: Beyond Journey's End",
							"native": "葬送のフリーレン",
							"userPreferred": "Frieren: Beyond Journey's End"
						},
						"coverImage": {
							"extraLarge": "https://img/anime-cover.jpg",
							"large": "https://img/anime-cover-large.jpg",
							"medium": "https://img/anime-cover-medium.jpg",
							"color": "#123456"
						},
						"bannerImage": "https://img/anime-banner.jpg",
						"startDate": {"year": 2023, "month": 9, "day": 29},
						"endDate": {"year": 2024, "month": 3, "day": 22},
						"studios": {
							"nodes": [{
								"id": 11,
								"name": "Madhouse",
								"isAnimationStudio": true
							}]
						},
						"streamingEpisodes": [{
							"title": "Episode 1",
							"thumbnail": "https://img/ep1.jpg",
							"url": "https://stream/ep1",
							"site": "Crunchyroll"
						}],
						"nextAiringEpisode": null
					}]
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewWithHTTPClient("", server.Client())
	client.baseURL = server.URL

	page, err := client.SearchAnime(context.Background(), "Frieren", 1, 10)
	if err != nil {
		t.Fatalf("SearchAnime error: %v", err)
	}
	if len(page.Media) != 1 {
		t.Fatalf("len(page.Media) = %d, want 1", len(page.Media))
	}

	got := page.Media[0]
	if got.PreferredTitle() != "Frieren: Beyond Journey's End" {
		t.Fatalf("PreferredTitle = %q", got.PreferredTitle())
	}
	if got.PrimaryStudio() != "Madhouse" {
		t.Fatalf("PrimaryStudio = %q, want Madhouse", got.PrimaryStudio())
	}
	if got.StartDate.String() != "2023-09-29" {
		t.Fatalf("StartDate.String() = %q, want 2023-09-29", got.StartDate.String())
	}
}

func TestGetAnimeByIDRetriesRateLimit(t *testing.T) {
	var hits atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := hits.Add(1)
		if current == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Too Many Requests.","status":429}]}`))
			return
		}

		_, _ = w.Write([]byte(`{
			"data": {
				"Media": {
					"id": 1,
					"type": "ANIME",
					"format": "TV",
					"status": "RELEASING",
					"episodes": 12,
					"description": "A test anime.",
					"siteUrl": "https://anilist.co/anime/1",
					"title": {
						"romaji": "Test Anime",
						"english": "Test Anime",
						"native": "テストアニメ",
						"userPreferred": "Test Anime"
					},
					"coverImage": {
						"extraLarge": "https://img/cover.jpg",
						"large": "https://img/cover-large.jpg",
						"medium": "https://img/cover-medium.jpg",
						"color": "#abcdef"
					},
					"bannerImage": "https://img/banner.jpg",
					"startDate": {"year": 2026, "month": 4, "day": 1},
					"endDate": {"year": 0, "month": 0, "day": 0},
					"studios": {"nodes": []},
					"streamingEpisodes": [],
					"nextAiringEpisode": {
						"airingAt": 1900000000,
						"timeUntilAiring": 3600,
						"episode": 7
					}
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewWithHTTPClient("", server.Client())
	client.baseURL = server.URL

	media, err := client.GetAnimeByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetAnimeByID error: %v", err)
	}
	if hits.Load() != 2 {
		t.Fatalf("request count = %d, want 2", hits.Load())
	}
	if media == nil || media.ID != 1 {
		t.Fatalf("media = %#v, want id 1", media)
	}
	if media.NextAiringEpisode == nil || media.NextAiringEpisode.Episode != 7 {
		t.Fatalf("next airing episode = %#v, want episode 7", media.NextAiringEpisode)
	}
}
