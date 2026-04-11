package com.velox.app.data.api

import com.velox.app.data.model.*
import retrofit2.Response
import retrofit2.http.*

interface VeloxApi {
    // Health check
    @GET("health")
    suspend fun healthCheck(): Response<MessageResponse>

    // Auth
    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): Response<DataWrapper<LoginResponse>>

    @POST("auth/refresh")
    suspend fun refreshToken(@Body request: RefreshRequest): Response<DataWrapper<RefreshResponse>>

    @POST("auth/logout")
    suspend fun logout(): Response<MessageResponse>

    @POST("auth/change-password")
    suspend fun changePassword(@Body request: ChangePasswordRequest): Response<MessageResponse>

    @GET("auth/me")
    suspend fun getMe(): Response<DataWrapper<UserDto>>

    @GET("auth/sessions")
    suspend fun getSessions(): Response<DataWrapper<List<SessionDto>>>

    @DELETE("auth/sessions/{id}")
    suspend fun revokeSession(@Path("id") sessionId: Int): Response<MessageResponse>

    // Media
    @GET("media")
    suspend fun getMediaList(
        @Query("type") type: String? = null,
        @Query("genre") genre: String? = null,
        @Query("year") year: String? = null,
        @Query("min_rating") minRating: Float? = null,
        @Query("sort") sort: String? = null,
        @Query("search") search: String? = null,
        @Query("library_id") libraryId: Int? = null,
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0,
    ): Response<DataWrapper<List<MediaListItemDto>>>

    @GET("media/{id}")
    suspend fun getMedia(@Path("id") mediaId: Int): Response<DataWrapper<MediaDto>>

    @GET("media/{id}/files")
    suspend fun getMediaWithFiles(@Path("id") mediaId: Int): Response<DataWrapper<MediaWithFilesDto>>

    @GET("media/{id}/credits")
    suspend fun getMediaCredits(@Path("id") mediaId: Int): Response<DataWrapper<CreditsDto>>

    // Series
    @GET("series")
    suspend fun getSeriesList(
        @Query("genre") genre: String? = null,
        @Query("year") year: String? = null,
        @Query("sort") sort: String? = null,
        @Query("search") search: String? = null,
        @Query("library_id") libraryId: Int? = null,
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0,
    ): Response<DataWrapper<List<SeriesListItemDto>>>

    @GET("series/{id}")
    suspend fun getSeries(@Path("id") seriesId: Int): Response<DataWrapper<SeriesDto>>

    @GET("series/{id}/seasons")
    suspend fun getSeasons(@Path("id") seriesId: Int): Response<DataWrapper<List<SeasonDto>>>

    @GET("series/{seriesId}/seasons/{seasonId}/episodes")
    suspend fun getEpisodes(
        @Path("seriesId") seriesId: Int,
        @Path("seasonId") seasonId: Int,
    ): Response<DataWrapper<List<EpisodeDto>>>

    @GET("series/{id}/credits")
    suspend fun getSeriesCredits(@Path("id") seriesId: Int): Response<DataWrapper<CreditsDto>>

    // Browse / Libraries
    @GET("browse")
    suspend fun browse(
        @Query("library_id") libraryId: Int? = null,
        @Query("path") path: String? = null,
    ): Response<DataWrapper<BrowseResponse>>

    @GET("libraries")
    suspend fun getLibraries(): Response<DataWrapper<List<LibraryDto>>>

    @GET("genres")
    suspend fun getGenres(@Query("type") type: String? = null): Response<DataWrapper<List<GenreDto>>>

    @GET("search")
    suspend fun search(
        @Query("q") query: String,
        @Query("limit") limit: Int = 20,
    ): Response<DataWrapper<SearchResponse>>

    // Playback
    @POST("playback/{id}/info")
    suspend fun getPlaybackInfo(
        @Path("id") mediaId: Int,
        @Body request: PlaybackInfoRequest? = null,
    ): Response<DataWrapper<PlaybackInfoDto>>

    @POST("stream/{id}/url")
    suspend fun getStreamUrl(@Path("id") mediaId: Int): Response<DataWrapper<StreamUrlResponse>>

    // Profile
    @GET("profile")
    suspend fun getProfile(): Response<DataWrapper<UserDto>>

    @PATCH("profile")
    suspend fun updateProfile(@Body updates: Map<String, String>): Response<DataWrapper<UserDto>>

    @GET("profile/preferences")
    suspend fun getPreferences(): Response<DataWrapper<UserPreferencesDto>>

    @PUT("profile/preferences")
    suspend fun updatePreferences(@Body request: UpdatePreferencesRequest): Response<DataWrapper<UserPreferencesDto>>

    @GET("profile/progress/{mediaId}")
    suspend fun getProgress(@Path("mediaId") mediaId: Int): Response<DataWrapper<UserDataDto?>>

    @PUT("profile/progress/{mediaId}")
    suspend fun updateProgress(
        @Path("mediaId") mediaId: Int,
        @Body request: UpdateProgressRequest,
    ): Response<MessageResponse>

    @DELETE("profile/progress/{mediaId}/dismiss")
    suspend fun dismissProgress(@Path("mediaId") mediaId: Int): Response<MessageResponse>

    @GET("profile/favorites")
    suspend fun getFavorites(
        @Query("limit") limit: Int = 20,
        @Query("offset") offset: Int = 0,
    ): Response<DataWrapper<List<MediaListItemDto>>>

    @POST("profile/favorites/{mediaId}")
    suspend fun toggleFavorite(@Path("mediaId") mediaId: Int): Response<ToggleFavoriteResponse>

    @GET("profile/recently-watched")
    suspend fun getRecentlyWatched(
        @Query("limit") limit: Int = 20,
    ): Response<DataWrapper<List<MediaListItemDto>>>

    @GET("profile/continue-watching")
    suspend fun getContinueWatching(
        @Query("limit") limit: Int = 20,
    ): Response<DataWrapper<List<ContinueWatchingItemDto>>>

    @GET("profile/next-up")
    suspend fun getNextUp(
        @Query("limit") limit: Int = 20,
    ): Response<DataWrapper<List<NextUpItemDto>>>

    // Notifications
    @GET("notifications")
    suspend fun getNotifications(
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0,
    ): Response<DataWrapper<NotificationListResponse>>

    @PATCH("notifications/read")
    suspend fun markNotificationsRead(
        @Body request: MarkReadRequest,
    ): Response<MessageResponse>

    @PATCH("notifications/read-all")
    suspend fun markAllNotificationsRead(): Response<MessageResponse>

    @POST("notifications/delete")
    suspend fun deleteNotifications(
        @Body request: DeleteNotificationsRequest,
    ): Response<MessageResponse>

    @GET("notifications/unread-count")
    suspend fun getUnreadCount(): Response<DataWrapper<UnreadCountResponse>>

    // Media Metadata
    @GET("media/{id}/genres")
    suspend fun getMediaGenres(
        @Path("id") mediaId: Int,
    ): Response<DataWrapper<List<GenreDto>>>

    @GET("media/{id}/cinema")
    suspend fun getMediaCinema(
        @Path("id") mediaId: Int,
    ): Response<DataWrapper<CinemaDto>>

    @POST("media/{id}/refresh")
    suspend fun refreshMediaMetadata(
        @Path("id") mediaId: Int,
    ): Response<MessageResponse>

    @PATCH("media/{id}/metadata")
    suspend fun updateMediaMetadata(
        @Path("id") mediaId: Int,
        @Body request: UpdateMetadataRequest,
    ): Response<DataWrapper<MediaDto>>

    @POST("media/{id}/images")
    suspend fun updateMediaImages(
        @Path("id") mediaId: Int,
        @Body request: UpdateImagesRequest,
    ): Response<MessageResponse>

    // Series Metadata
    @GET("series/{id}/genres")
    suspend fun getSeriesGenres(
        @Path("id") seriesId: Int,
    ): Response<DataWrapper<List<GenreDto>>>

    @GET("series/{id}/cinema")
    suspend fun getSeriesCinema(
        @Path("id") seriesId: Int,
    ): Response<DataWrapper<CinemaDto>>

    // Admin - Server
    @GET("admin/server")
    suspend fun getServerInfo(): Response<DataWrapper<ServerInfoDto>>

    @GET("admin/stats/libraries")
    suspend fun getLibraryStats(): Response<DataWrapper<List<LibraryStatsDto>>>

    // Admin - Libraries
    @POST("admin/libraries")
    suspend fun createLibrary(
        @Body request: CreateLibraryRequest,
    ): Response<DataWrapper<LibraryDto>>

    @PUT("admin/libraries/{id}")
    suspend fun updateLibrary(
        @Path("id") libraryId: Int,
        @Body request: UpdateLibraryRequest,
    ): Response<DataWrapper<LibraryDto>>

    @DELETE("admin/libraries/{id}")
    suspend fun deleteLibrary(
        @Path("id") libraryId: Int,
    ): Response<MessageResponse>

    @POST("admin/libraries/{id}/scan")
    suspend fun scanLibrary(
        @Path("id") libraryId: Int,
        @Query("force") force: Boolean? = null,
    ): Response<MessageResponse>

    // Admin - Users
    @GET("users")
    suspend fun getAdminUsers(): Response<DataWrapper<List<AdminUserDto>>>

    @POST("users")
    suspend fun createUser(
        @Body request: CreateUserRequest,
    ): Response<DataWrapper<AdminUserDto>>

    @PATCH("users/{id}")
    suspend fun updateUser(
        @Path("id") userId: Int,
        @Body request: UpdateUserRequest,
    ): Response<DataWrapper<AdminUserDto>>

    @DELETE("users/{id}")
    suspend fun deleteUser(
        @Path("id") userId: Int,
    ): Response<MessageResponse>

    // Admin - Activity
    @GET("admin/activity")
    suspend fun getAdminActivity(
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0,
    ): Response<DataWrapper<List<ActivityLogDto>>>

    // Admin - Tasks
    @GET("admin/tasks")
    suspend fun getTasks(): Response<DataWrapper<List<TaskDto>>>

    @POST("admin/tasks/{name}/run")
    suspend fun runTask(
        @Path("name") name: String,
    ): Response<MessageResponse>

    // Admin - Webhooks
    @GET("admin/webhooks")
    suspend fun getWebhooks(): Response<DataWrapper<List<WebhookDto>>>

    @POST("admin/webhooks")
    suspend fun createWebhook(
        @Body request: CreateWebhookRequest,
    ): Response<DataWrapper<WebhookDto>>

    @PUT("admin/webhooks/{id}")
    suspend fun updateWebhook(
        @Path("id") webhookId: Int,
        @Body request: UpdateWebhookRequest,
    ): Response<DataWrapper<WebhookDto>>

    @DELETE("admin/webhooks/{id}")
    suspend fun deleteWebhook(
        @Path("id") webhookId: Int,
    ): Response<MessageResponse>

    @POST("admin/webhooks/{id}/test")
    suspend fun testWebhook(
        @Path("id") webhookId: Int,
    ): Response<DataWrapper<WebhookTestResponse>>

    // Admin - Settings
    @GET("admin/settings")
    suspend fun getAdminSettings(): Response<DataWrapper<AdminSettingsDto>>

    @PUT("admin/settings")
    suspend fun updateAdminSettings(
        @Body request: UpdateAdminSettingsRequest,
    ): Response<DataWrapper<AdminSettingsDto>>

    // Provider Settings
    @GET("admin/settings/{provider}")
    suspend fun getProviderSettings(@Path("provider") provider: String): Response<DataWrapper<ProviderSettingsDto>>

    @PUT("admin/settings/{provider}")
    suspend fun updateProviderSettings(
        @Path("provider") provider: String,
        @Body request: UpdateProviderRequest,
    ): Response<DataWrapper<ProviderSettingsDto>>

    // OpenSubtitles Settings
    @GET("admin/settings/opensubtitles")
    suspend fun getOpenSubsSettings(): Response<DataWrapper<OpenSubsSettingsDto>>

    @PUT("admin/settings/opensubtitles")
    suspend fun updateOpenSubsSettings(
        @Body request: UpdateOpenSubsRequest,
    ): Response<DataWrapper<OpenSubsSettingsDto>>

    @GET("admin/settings/ai-translation")
    suspend fun getAITranslationSettings(): Response<DataWrapper<AITranslationSettingsDto>>

    @PUT("admin/settings/ai-translation")
    suspend fun updateAITranslationSettings(
        @Body request: UpdateAITranslationRequest,
    ): Response<DataWrapper<AITranslationSettingsDto>>

    // Auto-Subtitle Settings
    @GET("admin/settings/auto-subtitles")
    suspend fun getAutoSubSettings(): Response<DataWrapper<AutoSubSettingsDto>>

    @PUT("admin/settings/auto-subtitles")
    suspend fun updateAutoSubSettings(
        @Body request: UpdateAutoSubRequest,
    ): Response<DataWrapper<AutoSubSettingsDto>>

    // Playback Settings
    @GET("admin/settings/playback")
    suspend fun getPlaybackSettings(): Response<DataWrapper<PlaybackSettingsDto>>

    @PUT("admin/settings/playback")
    suspend fun updatePlaybackSettings(
        @Body request: UpdatePlaybackRequest,
    ): Response<DataWrapper<PlaybackSettingsDto>>

    // Cinema Settings
    @GET("admin/settings/cinema")
    suspend fun getCinemaSettings(): Response<DataWrapper<CinemaSettingsDto>>

    @PUT("admin/settings/cinema")
    suspend fun updateCinemaSettings(
        @Body request: UpdateCinemaRequest,
    ): Response<DataWrapper<CinemaSettingsDto>>

    // Pre-transcode Settings & Logic
    @GET("admin/settings/pretranscode")
    suspend fun getPretranscodeSettings(): Response<DataWrapper<PretranscodeSettingsDto>>

    @PUT("admin/settings/pretranscode")
    suspend fun updatePretranscodeSettings(
        @Body request: UpdatePretranscodeRequest,
    ): Response<DataWrapper<PretranscodeSettingsDto>>

    // Admin Markers Dashboard
    @GET("admin/markers/stats")
    suspend fun getMarkerStats(): Response<DataWrapper<MarkerStatsDto>>

    @POST("admin/markers/backfill")
    suspend fun startMarkerBackfill(
        @Body request: BackfillMarkersRequest,
    ): Response<DataWrapper<BackfillResultDto>>

    @GET("admin/pretranscode/status")
    suspend fun getPretranscodeStatus(): Response<DataWrapper<PretranscodeStatusDto>>

    @GET("admin/pretranscode/profiles")
    suspend fun getPretranscodeProfiles(): Response<DataWrapper<List<PretranscodeProfileDto>>>

    @PUT("admin/pretranscode/profiles/{id}")
    suspend fun togglePretranscodeProfile(
        @Path("id") id: Int,
        @Body request: Map<String, Boolean>,
    ): Response<DataWrapper<Any>>

    @GET("admin/pretranscode/estimate")
    suspend fun getPretranscodeEstimate(
        @Query("library_id") libraryId: Int,
    ): Response<DataWrapper<StorageEstimateDto>>

    @POST("admin/pretranscode/start")
    suspend fun startPretranscode(): Response<DataWrapper<Any>>

    @POST("admin/pretranscode/stop")
    suspend fun stopPretranscode(): Response<DataWrapper<Any>>

    @POST("admin/pretranscode/resume")
    suspend fun resumePretranscode(): Response<DataWrapper<Any>>

    @POST("admin/pretranscode/cleanup-files")
    suspend fun cleanupPretranscodeFiles(): Response<DataWrapper<Any>>

    // Metadata Refresh
    @POST("admin/metadata/refresh-ratings")
    suspend fun refreshAllMetadata(): Response<DataWrapper<Map<String, Int>>>

    // Subtitles - Search
    @GET("media/{id}/subtitles/search")
    suspend fun searchSubtitles(
        @Path("id") mediaId: Int,
        @Query("lang") lang: String,
    ): Response<DataWrapper<List<SubtitleSearchResultDto>>>

    // Subtitles - Download
    @POST("media/{id}/subtitles/download")
    suspend fun downloadSubtitle(
        @Path("id") mediaId: Int,
        @Body request: SubtitleDownloadRequest,
    ): Response<DataWrapper<SubtitleDto>>

    // Subtitles - Translate
    @POST("subtitles/{id}/translate")
    suspend fun translateSubtitle(
        @Path("id") subtitleId: Int,
        @Body request: TranslateSubtitleRequest,
    ): Response<DataWrapper<SubtitleDto>>

    // Subtitles - Serve Raw Content
    @GET("media-files/{fileId}/subtitles/{subId}/serve")
    suspend fun getSubtitleContent(
        @Path("fileId") fileId: Int,
        @Path("subId") subId: Int,
    ): Response<okhttp3.ResponseBody>
}
