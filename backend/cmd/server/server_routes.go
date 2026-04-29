package main

import (
	"context"
	"net/http"
	"time"

	"github.com/thawng/velox/internal/handler"
	"github.com/thawng/velox/internal/middleware"
)

var authSkipPaths = []string{
	"/api/health",
	"/api/setup/status",
	"/api/setup",
	"/api/setup/wizard",
	"/api/auth/login",
	"/api/auth/refresh",
	"/api/auth/logout",
	"/api/images/*",
	"/api/media/*/trickplay/*",
	"/api/app-versions/latest",
	// Live TV stream proxy: <video> cannot send Authorization headers. Both
	// endpoints are self-contained — stream entry only leaks already-public
	// channel URLs; proxy URLs are HMAC-signed with short expiry.
	"/api/livetv/stream/*",
	"/api/livetv/proxy",
}

func (app *serverApp) newHTTPServer() *http.Server {
	return &http.Server{
		Addr:         app.cfg.Addr(),
		Handler:      app.newHTTPHandler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}
}

func (app *serverApp) newHTTPHandler() http.Handler {
	sessionTracker := middleware.SessionTracker(func(sessionID int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.repos.session.UpdateLastActive(ctx, sessionID)
	})

	authMiddleware := middleware.RequireAuth(app.jwtManager, app.apiKeyStore, authSkipPaths...)

	var h http.Handler = app.newMux()
	h = sessionTracker(h)
	h = authMiddleware(h)
	h = middleware.CORS(app.cfg.CORSOrigins)(h)
	h = middleware.Logger(h)
	h = middleware.Recovery(h)

	return h
}

func (app *serverApp) newMux() *http.ServeMux {
	mux := http.NewServeMux()

	app.registerHealthRoutes(mux)
	app.registerSetupRoutes(mux)
	app.registerAuthRoutes(mux)
	app.registerAdminUserRoutes(mux)
	app.registerAdminSettingsRoutes(mux)
	app.registerAdminOperationsRoutes(mux)
	app.registerLibraryRoutes(mux)
	app.registerProfileRoutes(mux)
	app.registerDiscoveryRoutes(mux)
	app.registerMediaRoutes(mux)
	app.registerSeriesRoutes(mux)
	app.registerMetadataRoutes(mux)
	app.registerCinemaRoutes(mux)
	app.registerStreamRoutes(mux)
	app.registerPlaybackRoutes(mux)
	app.registerSubtitleRoutes(mux)
	app.registerNotificationRoutes(mux)
	app.registerAudioTrackRoutes(mux)
	app.registerAppVersionRoutes(mux)
	app.registerDownloadRoutes(mux)
	app.registerLiveTVRoutes(mux)

	return mux
}

func (app *serverApp) registerLiveTVRoutes(mux *http.ServeMux) {
	app.handlers.liveTV.RegisterRoutes(mux)
}

func (app *serverApp) registerDownloadRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/media/{id}/download", middleware.RequireAdmin(http.HandlerFunc(app.handlers.download.StartDownload)))
	mux.Handle("POST /api/series/{id}/download", middleware.RequireAdmin(http.HandlerFunc(app.handlers.download.StartSeriesDownload)))
	mux.Handle("DELETE /api/series/{id}/download", middleware.RequireAdmin(http.HandlerFunc(app.handlers.download.DeleteSeriesDownload)))
	mux.Handle("DELETE /api/media/{id}/download", middleware.RequireAdmin(http.HandlerFunc(app.handlers.download.DeleteDownload)))
	mux.Handle("GET /api/downloads", middleware.RequireAdmin(http.HandlerFunc(app.handlers.download.GetTasks)))
}

func (app *serverApp) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", handler.Health)
}

func (app *serverApp) registerSetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/setup/status", app.handlers.setup.Status)
	mux.HandleFunc("POST /api/setup", app.handlers.setup.Setup)
	mux.HandleFunc("GET /api/setup/wizard", app.handlers.setup.WizardStatus)
	mux.HandleFunc("POST /api/setup/wizard/complete", app.handlers.setup.CompleteWizard)
}

func (app *serverApp) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", app.handlers.auth.Login)
	mux.HandleFunc("POST /api/auth/refresh", app.handlers.auth.Refresh)
	mux.HandleFunc("POST /api/auth/logout", app.handlers.auth.Logout)
	mux.HandleFunc("POST /api/auth/change-password", app.handlers.auth.ChangePassword)
	mux.HandleFunc("GET /api/auth/me", app.handlers.auth.Me)
	mux.HandleFunc("GET /api/auth/sessions", app.handlers.auth.ListSessions)
	mux.HandleFunc("DELETE /api/auth/sessions/{id}", app.handlers.auth.RevokeSession)
}

func (app *serverApp) registerAdminUserRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/users", middleware.RequireAdmin(http.HandlerFunc(app.handlers.user.List)))
	mux.Handle("POST /api/users", middleware.RequireAdmin(http.HandlerFunc(app.handlers.user.Create)))
	mux.Handle("PATCH /api/users/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.user.Update)))
	mux.Handle("DELETE /api/users/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.user.Delete)))
	mux.Handle("PUT /api/users/{id}/library-access", middleware.RequireAdmin(http.HandlerFunc(app.handlers.user.SetLibraryAccess)))
	mux.Handle("GET /api/admin/fs/browse", middleware.RequireAdmin(http.HandlerFunc(handler.FSBrowse)))
}

func (app *serverApp) registerAdminSettingsRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/settings/opensubtitles", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetOpenSubtitles)))
	mux.Handle("PUT /api/admin/settings/opensubtitles", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateOpenSubtitles)))
	mux.Handle("GET /api/admin/settings/anilist", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetAniList)))
	mux.Handle("PUT /api/admin/settings/anilist", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateAniList)))
	mux.Handle("GET /api/admin/settings/tmdb", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetTMDb)))
	mux.Handle("PUT /api/admin/settings/tmdb", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateTMDb)))
	mux.Handle("GET /api/admin/settings/omdb", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetOMDb)))
	mux.Handle("PUT /api/admin/settings/omdb", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateOMDb)))
	mux.Handle("GET /api/admin/settings/tvdb", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetTVDB)))
	mux.Handle("PUT /api/admin/settings/tvdb", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateTVDB)))
	mux.Handle("GET /api/admin/settings/fanart", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetFanart)))
	mux.Handle("PUT /api/admin/settings/fanart", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateFanart)))
	mux.Handle("GET /api/admin/settings/subdl", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetSubdl)))
	mux.Handle("PUT /api/admin/settings/subdl", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateSubdl)))
	mux.Handle("GET /api/admin/settings/deepl", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetDeepL)))
	mux.Handle("PUT /api/admin/settings/deepl", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateDeepL)))
	mux.Handle("GET /api/admin/settings/ai-translation", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetAITranslation)))
	mux.Handle("PUT /api/admin/settings/ai-translation", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateAITranslation)))
	mux.Handle("GET /api/admin/settings/auto-subtitles", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetAutoSubtitles)))
	mux.Handle("PUT /api/admin/settings/auto-subtitles", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateAutoSubtitles)))
	mux.Handle("GET /api/admin/settings/auto-translate", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetAutoTranslate)))
	mux.Handle("PUT /api/admin/settings/auto-translate", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdateAutoTranslate)))
	mux.Handle("GET /api/admin/settings/playback", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.GetPlayback)))
	mux.Handle("PUT /api/admin/settings/playback", middleware.RequireAdmin(http.HandlerFunc(app.handlers.settings.UpdatePlayback)))
	mux.Handle("GET /api/admin/settings/cinema", middleware.RequireAdmin(http.HandlerFunc(app.handlers.cinema.GetSettings)))
	mux.Handle("PUT /api/admin/settings/cinema", middleware.RequireAdmin(http.HandlerFunc(app.handlers.cinema.UpdateSettings)))
}

func (app *serverApp) registerAdminOperationsRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/activity", middleware.RequireAdmin(http.HandlerFunc(app.handlers.activity.List)))
	mux.Handle("GET /api/admin/stats/playback", middleware.RequireAdmin(http.HandlerFunc(app.handlers.activity.GetPlaybackStats)))
	mux.Handle("GET /api/admin/server", middleware.RequireAdmin(http.HandlerFunc(app.handlers.admin.ServerInfo)))
	mux.Handle("GET /api/admin/stats/libraries", middleware.RequireAdmin(http.HandlerFunc(app.handlers.admin.LibraryStats)))
	mux.Handle("POST /api/admin/ophim/sync", middleware.RequireAdmin(http.HandlerFunc(app.handlers.admin.SyncOphim)))
	mux.Handle("GET /api/admin/webhooks", middleware.RequireAdmin(http.HandlerFunc(app.handlers.webhook.List)))
	mux.Handle("POST /api/admin/webhooks", middleware.RequireAdmin(http.HandlerFunc(app.handlers.webhook.Create)))
	mux.Handle("PUT /api/admin/webhooks/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.webhook.Update)))
	mux.Handle("DELETE /api/admin/webhooks/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.webhook.Delete)))
	mux.Handle("GET /api/admin/markers/stats", middleware.RequireAdmin(http.HandlerFunc(app.handlers.markerAdmin.Stats)))
	mux.Handle("GET /api/admin/markers/detectors", middleware.RequireAdmin(http.HandlerFunc(app.handlers.markerAdmin.ListDetectors)))
	mux.Handle("POST /api/admin/markers/detect", middleware.RequireAdmin(http.HandlerFunc(app.handlers.markerAdmin.DetectWithDetector)))
	mux.Handle("POST /api/admin/markers/backfill", middleware.RequireAdmin(http.HandlerFunc(app.handlers.markerAdmin.BackfillMarkers)))
	mux.Handle("GET /api/admin/pretranscode/status", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.GetStatus)))
	mux.Handle("POST /api/admin/pretranscode/start", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.Start)))
	mux.Handle("POST /api/admin/pretranscode/stop", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.Stop)))
	mux.Handle("POST /api/admin/pretranscode/resume", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.Resume)))
	mux.Handle("GET /api/admin/pretranscode/estimate", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.Estimate)))
	mux.Handle("POST /api/admin/pretranscode/cleanup-files", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.Cleanup)))
	mux.Handle("GET /api/admin/pretranscode/profiles", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.ListProfiles)))
	mux.Handle("PUT /api/admin/pretranscode/profiles/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.ToggleProfile)))
	mux.Handle("GET /api/admin/settings/pretranscode", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.GetSettings)))
	mux.Handle("PUT /api/admin/settings/pretranscode", middleware.RequireAdmin(http.HandlerFunc(app.handlers.pretranscode.UpdateSettings)))
	mux.Handle("GET /api/admin/tasks", middleware.RequireAdmin(http.HandlerFunc(app.handlers.scheduler.ListTasks)))
	mux.Handle("POST /api/admin/tasks/{name}/run", middleware.RequireAdmin(http.HandlerFunc(app.handlers.scheduler.RunTask)))
	mux.Handle("PATCH /api/admin/tasks/{name}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.scheduler.UpdateTask)))

	// Cloud storage providers (Plan W)
	mux.Handle("GET /api/admin/cloud/drivers", middleware.RequireAdmin(http.HandlerFunc(app.handlers.storageProvider.ListDrivers)))
	mux.Handle("GET /api/admin/cloud/providers", middleware.RequireAdmin(http.HandlerFunc(app.handlers.storageProvider.List)))
	mux.Handle("POST /api/admin/cloud/providers", middleware.RequireAdmin(http.HandlerFunc(app.handlers.storageProvider.Create)))
	mux.Handle("DELETE /api/admin/cloud/providers/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.storageProvider.Delete)))
	mux.Handle("POST /api/admin/cloud/providers/{id}/refresh", middleware.RequireAdmin(http.HandlerFunc(app.handlers.storageProvider.Refresh)))
	mux.Handle("POST /api/admin/cloud/providers/{id}/validate-url", middleware.RequireAdmin(http.HandlerFunc(app.handlers.storageProvider.ValidateURL)))
}

func (app *serverApp) registerLibraryRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/libraries", middleware.RequireAdmin(http.HandlerFunc(app.handlers.library.Create)))
	mux.Handle("DELETE /api/libraries/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.library.Delete)))
	mux.Handle("POST /api/libraries/{id}/scan", middleware.RequireAdmin(http.HandlerFunc(app.handlers.library.Scan)))
	mux.Handle("GET /api/libraries/{id}/scan-status", middleware.RequireAdmin(http.HandlerFunc(app.handlers.library.ScanStatus)))
	mux.HandleFunc("GET /api/libraries", app.handlers.library.List)
}

func (app *serverApp) registerProfileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/profile", app.handlers.profile.GetProfile)
	mux.HandleFunc("PATCH /api/profile", app.handlers.profile.UpdateProfile)
	mux.HandleFunc("GET /api/profile/preferences", app.handlers.profile.GetPreferences)
	mux.HandleFunc("PUT /api/profile/preferences", app.handlers.profile.UpdatePreferences)
	mux.HandleFunc("GET /api/profile/progress/{mediaId}", app.handlers.profile.GetProgress)
	mux.HandleFunc("PUT /api/profile/progress/{mediaId}", app.handlers.profile.UpdateProgress)
	mux.HandleFunc("GET /api/profile/favorites", app.handlers.profile.ListFavorites)
	mux.HandleFunc("GET /api/profile/favorites/alphabet", app.handlers.profile.GetFavoritesAlphabet)
	mux.HandleFunc("POST /api/profile/favorites/{mediaId}", app.handlers.profile.ToggleFavorite)
	mux.HandleFunc("GET /api/profile/recently-watched", app.handlers.profile.ListRecentlyWatched)
	mux.HandleFunc("GET /api/profile/continue-watching", app.handlers.profile.ContinueWatching)
	mux.HandleFunc("GET /api/profile/next-up", app.handlers.profile.NextUp)
	mux.HandleFunc("DELETE /api/profile/progress/{mediaId}/dismiss", app.handlers.profile.DismissProgress)
}

func (app *serverApp) registerDiscoveryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/browse", app.handlers.browse.Browse)
	mux.HandleFunc("GET /api/genres", app.handlers.catalog.ListGenres)
	mux.HandleFunc("GET /api/media/{id}/genres", app.handlers.catalog.ListMediaGenres)
	mux.HandleFunc("GET /api/media/{id}/credits", app.handlers.catalog.ListMediaCredits)
	mux.HandleFunc("GET /api/series/{id}/genres", app.handlers.catalog.ListSeriesGenres)
	mux.HandleFunc("GET /api/series/{id}/credits", app.handlers.catalog.ListSeriesCredits)
	mux.HandleFunc("GET /api/search", app.handlers.catalog.Search)
}

func (app *serverApp) registerMediaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/media", app.handlers.media.List)
	mux.HandleFunc("GET /api/media/alphabet", app.handlers.media.GetAlphabet)
	mux.HandleFunc("GET /api/media/{id}", app.handlers.media.Get)
	mux.HandleFunc("GET /api/media/{id}/files", app.handlers.media.GetWithFiles)
	mux.HandleFunc("GET /api/media/{id}/versions", app.handlers.media.GetVersions)
}

func (app *serverApp) registerSeriesRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/series", app.handlers.series.ListSeries)
	mux.HandleFunc("GET /api/series/alphabet", app.handlers.series.GetAlphabet)
	mux.HandleFunc("GET /api/series/search", app.handlers.series.SearchSeries)
	mux.HandleFunc("GET /api/series/{id}", app.handlers.series.GetSeries)
	mux.HandleFunc("GET /api/series/{id}/seasons", app.handlers.series.ListSeasons)
	mux.HandleFunc("GET /api/series/{id}/seasons/{seasonId}/episodes", app.handlers.series.ListEpisodes)
}

func (app *serverApp) registerMetadataRoutes(mux *http.ServeMux) {
	if app.handlers.metadata == nil {
		return
	}

	mux.Handle("PUT /api/media/{id}/identify", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.Identify)))
	mux.Handle("POST /api/media/{id}/refresh", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.Refresh)))
	mux.Handle("POST /api/media/{id}/cloud-probe", middleware.RequireAdmin(http.HandlerFunc(app.handlers.streamURL.CloudProbe)))
	mux.Handle("POST /api/admin/metadata/refresh-ratings", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.BulkRefreshRatings)))
	mux.Handle("POST /api/admin/metadata/force-refresh", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.ForceBulkRefreshMetadata)))
	mux.Handle("PATCH /api/media/{id}/metadata", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.EditMediaMetadata)))
	mux.Handle("PATCH /api/series/{id}/metadata", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.EditSeriesMetadata)))
	mux.Handle("PATCH /api/episodes/{id}/metadata", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.EditEpisodeMetadata)))
	mux.Handle("DELETE /api/media/{id}/metadata/lock", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.UnlockMediaMetadata)))
	mux.Handle("DELETE /api/series/{id}/metadata/lock", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.UnlockSeriesMetadata)))
	mux.Handle("POST /api/media/{id}/images", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.UploadMediaImage)))
	mux.Handle("POST /api/series/{id}/images", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.UploadSeriesImage)))
	mux.Handle("DELETE /api/media/{id}/images/{imageType}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.DeleteMediaImage)))
	mux.Handle("DELETE /api/series/{id}/images/{imageType}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.DeleteSeriesImage)))
	mux.Handle("POST /api/media/{id}/nfo", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.WriteMediaNFO)))
	mux.Handle("POST /api/series/{id}/nfo", middleware.RequireAdmin(http.HandlerFunc(app.handlers.metadata.WriteSeriesNFO)))
	mux.HandleFunc("GET /api/images/local/{type}/{id}/{filename}", app.handlers.metadata.ServeLocalImage)
}

func (app *serverApp) registerCinemaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/media/{id}/cinema", app.handlers.cinema.MediaCinema)
	mux.HandleFunc("GET /api/series/{id}/cinema", app.handlers.cinema.SeriesCinema)
	mux.HandleFunc("GET /api/cinema/intro", app.handlers.cinema.ServeIntro)
	mux.Handle("POST /api/admin/cinema/intro", middleware.RequireAdmin(http.HandlerFunc(app.handlers.cinema.UploadIntro)))
}

func (app *serverApp) registerStreamRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/stream/{id}/url", app.handlers.streamURL.Create)
	mux.HandleFunc("GET /api/stream/{id}", app.handlers.stream.DirectPlay)
	mux.HandleFunc("GET /api/stream/{id}/hls/master.m3u8", app.handlers.stream.HLSMaster)
	mux.HandleFunc("GET /api/stream/{id}/hls/abr.m3u8", app.handlers.stream.HLSABRMaster)
	mux.HandleFunc("GET /api/stream/{id}/hls/{segment}", app.handlers.stream.HLSSegment)

	mux.HandleFunc("DELETE /api/stream/sessions/{ssid}", app.handlers.stream.StopTranscodeBySession)
	mux.HandleFunc("DELETE /api/stream/by-id/{id}/session", app.handlers.stream.StopTranscode)
	mux.HandleFunc("GET /api/media/{id}/trickplay/manifest.vtt", app.handlers.trickplay.ServeVTT)
	mux.HandleFunc("GET /api/media/{id}/trickplay/{sprite}", app.handlers.trickplay.ServeSprite)
	// Width selection moved to query param: /api/images/tmdb/{path...}?width=500
	mux.HandleFunc("GET /api/images/tmdb/{type}/{path...}", app.handlers.image.Serve)
}

func (app *serverApp) registerPlaybackRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/playback/{id}/info", app.handlers.playback.GetPlaybackInfo)
	mux.HandleFunc("GET /api/playback/capabilities", app.handlers.playback.GetClientCapabilities)
}

func (app *serverApp) registerSubtitleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/media-files/{media_file_id}/subtitles", app.handlers.subtitle.ListByMediaFile)
	mux.HandleFunc("GET /api/media-files/{media_file_id}/subtitles/{subtitle_id}/serve", app.handlers.subtitle.Serve)
	mux.Handle("POST /api/subtitles", middleware.RequireAdmin(http.HandlerFunc(app.handlers.subtitle.Create)))
	mux.Handle("PATCH /api/subtitles/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.subtitle.Update)))
	mux.Handle("DELETE /api/subtitles/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.subtitle.Delete)))
	mux.Handle("POST /api/media-files/{media_file_id}/subtitles/{subtitle_id}/default", middleware.RequireAdmin(http.HandlerFunc(app.handlers.subtitle.SetDefault)))
	mux.HandleFunc("POST /api/subtitles/{id}/translate", app.handlers.subtitle.Translate)
	mux.HandleFunc("GET /api/media/{id}/subtitles/search", app.handlers.subtitleSearch.Search)
	mux.HandleFunc("POST /api/media/{id}/subtitles/download", app.handlers.subtitleSearch.Download)
	mux.HandleFunc("POST /api/media/{id}/subtitles/auto-download", app.handlers.subtitleSearch.AutoDownload)
}

func (app *serverApp) registerNotificationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ws", app.handlers.ws.Handle)
	mux.HandleFunc("GET /api/notifications", app.handlers.notification.List)
	mux.HandleFunc("GET /api/notifications/unread-count", app.handlers.notification.UnreadCount)
	mux.HandleFunc("GET /api/notifications/types", app.handlers.notification.Types)
	mux.HandleFunc("PATCH /api/notifications/read", app.handlers.notification.MarkAsRead)
	mux.HandleFunc("PATCH /api/notifications/read-all", app.handlers.notification.MarkAllAsRead)
	mux.HandleFunc("DELETE /api/notifications/{id}", app.handlers.notification.DeleteOne)
	mux.HandleFunc("POST /api/notifications/delete", app.handlers.notification.Delete)
}

func (app *serverApp) registerAudioTrackRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/media-files/{media_file_id}/audio-tracks", app.handlers.audioTrack.ListByMediaFile)
	mux.Handle("POST /api/audio-tracks", middleware.RequireAdmin(http.HandlerFunc(app.handlers.audioTrack.Create)))
	mux.Handle("PATCH /api/audio-tracks/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.audioTrack.Update)))
	mux.Handle("DELETE /api/audio-tracks/{id}", middleware.RequireAdmin(http.HandlerFunc(app.handlers.audioTrack.Delete)))
	mux.Handle("POST /api/media-files/{media_file_id}/audio-tracks/{track_id}/default", middleware.RequireAdmin(http.HandlerFunc(app.handlers.audioTrack.SetDefault)))
}

func (app *serverApp) registerAppVersionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/app-versions/latest", app.handlers.appVersion.GetLatest)
}
