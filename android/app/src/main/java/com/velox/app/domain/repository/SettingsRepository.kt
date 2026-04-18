package com.velox.app.domain.repository

import com.velox.app.data.model.*
import retrofit2.Response

interface SettingsRepository {
    suspend fun healthCheck(): retrofit2.Response<MessageResponse>
    suspend fun refreshToken(request: RefreshRequest): retrofit2.Response<DataWrapper<RefreshResponse>>
    suspend fun logout(): retrofit2.Response<MessageResponse>
    suspend fun changePassword(request: ChangePasswordRequest): retrofit2.Response<MessageResponse>
    suspend fun getMe(): retrofit2.Response<DataWrapper<UserDto>>
    suspend fun getSessions(): retrofit2.Response<DataWrapper<List<SessionDto>>>
    suspend fun revokeSession(sessionId: Int): retrofit2.Response<MessageResponse>
    suspend fun getMediaList(type: String? = null, genre: String? = null, year: String? = null, minRating: Float? = null, sort: String? = null, search: String? = null, libraryId: Int? = null, limit: Int = 50, offset: Int = 0,): retrofit2.Response<DataWrapper<List<MediaListItemDto>>>
    suspend fun getMedia(mediaId: Int): retrofit2.Response<DataWrapper<MediaDto>>
    suspend fun getMediaWithFiles(mediaId: Int): retrofit2.Response<DataWrapper<MediaWithFilesDto>>
    suspend fun getMediaCredits(mediaId: Int): retrofit2.Response<DataWrapper<CreditsDto>>
    suspend fun getSeriesList(genre: String? = null, year: String? = null, sort: String? = null, search: String? = null, libraryId: Int? = null, limit: Int = 50, offset: Int = 0,): retrofit2.Response<DataWrapper<List<SeriesListItemDto>>>
    suspend fun getSeriesCredits(seriesId: Int): retrofit2.Response<DataWrapper<CreditsDto>>
    suspend fun browse(libraryId: Int? = null, path: String? = null,): retrofit2.Response<DataWrapper<BrowseResponse>>
    suspend fun getGenres(type: String? = null): retrofit2.Response<DataWrapper<List<GenreDto>>>
    suspend fun search(query: String, limit: Int = 20,): retrofit2.Response<DataWrapper<SearchResponse>>
    suspend fun getPlaybackInfo(mediaId: Int, request: PlaybackInfoRequest? = null,): retrofit2.Response<DataWrapper<PlaybackInfoDto>>
    suspend fun getStreamUrl(mediaId: Int): retrofit2.Response<DataWrapper<StreamUrlResponse>>
    suspend fun getProfile(): retrofit2.Response<DataWrapper<UserDto>>
    suspend fun updateProfile(updates: Map<String, String>): retrofit2.Response<DataWrapper<UserDto>>
    suspend fun getPreferences(): retrofit2.Response<DataWrapper<UserPreferencesDto>>
    suspend fun updatePreferences(request: UpdatePreferencesRequest): retrofit2.Response<DataWrapper<UserPreferencesDto>>
    suspend fun getProgress(mediaId: Int): retrofit2.Response<DataWrapper<UserDataDto?>>
    suspend fun updateProgress(mediaId: Int, request: UpdateProgressRequest,): retrofit2.Response<MessageResponse>
    suspend fun dismissProgress(mediaId: Int): retrofit2.Response<MessageResponse>
    suspend fun getFavorites(limit: Int = 20, offset: Int = 0,): retrofit2.Response<DataWrapper<List<MediaListItemDto>>>
    suspend fun toggleFavorite(mediaId: Int): retrofit2.Response<DataWrapper<ToggleFavoriteResponse>>
    suspend fun getRecentlyWatched(limit: Int = 20,): retrofit2.Response<DataWrapper<List<MediaListItemDto>>>
    suspend fun getContinueWatching(limit: Int = 20,): retrofit2.Response<DataWrapper<List<ContinueWatchingItemDto>>>
    suspend fun getNextUp(limit: Int = 20,): retrofit2.Response<DataWrapper<List<NextUpItemDto>>>
    suspend fun getNotifications(limit: Int = 50, offset: Int = 0,): retrofit2.Response<DataWrapper<NotificationListResponse>>
    suspend fun markNotificationsRead(request: MarkReadRequest,): retrofit2.Response<MessageResponse>
    suspend fun markAllNotificationsRead(): retrofit2.Response<MessageResponse>
    suspend fun deleteNotifications(request: DeleteNotificationsRequest,): retrofit2.Response<MessageResponse>
    suspend fun getUnreadCount(): retrofit2.Response<DataWrapper<UnreadCountResponse>>
    suspend fun getMediaGenres(mediaId: Int,): retrofit2.Response<DataWrapper<List<GenreDto>>>
    suspend fun getMediaCinema(mediaId: Int,): retrofit2.Response<DataWrapper<CinemaDto>>
    suspend fun refreshMediaMetadata(mediaId: Int,): retrofit2.Response<MessageResponse>
    suspend fun updateMediaMetadata(mediaId: Int, request: UpdateMetadataRequest,): retrofit2.Response<DataWrapper<MediaDto>>
    suspend fun updateMediaImages(mediaId: Int, request: UpdateImagesRequest,): retrofit2.Response<MessageResponse>
    suspend fun getSeriesGenres(seriesId: Int,): retrofit2.Response<DataWrapper<List<GenreDto>>>
    suspend fun getSeriesCinema(seriesId: Int,): retrofit2.Response<DataWrapper<CinemaDto>>
    suspend fun getServerInfo(): retrofit2.Response<DataWrapper<ServerInfoDto>>
    suspend fun getLatestAppVersion(platform: String = "android"): retrofit2.Response<AppVersionDto>
    suspend fun getLibraryStats(): retrofit2.Response<DataWrapper<List<LibraryStatsDto>>>
    suspend fun createLibrary(request: CreateLibraryRequest,): retrofit2.Response<DataWrapper<LibraryDto>>
    suspend fun updateLibrary(libraryId: Int, request: UpdateLibraryRequest,): retrofit2.Response<DataWrapper<LibraryDto>>
    suspend fun deleteLibrary(libraryId: Int,): retrofit2.Response<MessageResponse>
    suspend fun scanLibrary(libraryId: Int, force: Boolean? = null,): retrofit2.Response<MessageResponse>
    suspend fun getAdminUsers(): retrofit2.Response<DataWrapper<List<AdminUserDto>>>
    suspend fun createUser(request: CreateUserRequest,): retrofit2.Response<DataWrapper<AdminUserDto>>
    suspend fun updateUser(userId: Int, request: UpdateUserRequest,): retrofit2.Response<DataWrapper<AdminUserDto>>
    suspend fun deleteUser(userId: Int,): retrofit2.Response<MessageResponse>
    suspend fun getAdminActivity(limit: Int = 50, offset: Int = 0,): retrofit2.Response<DataWrapper<List<ActivityLogDto>>>
    suspend fun getTasks(): retrofit2.Response<DataWrapper<List<TaskDto>>>
    suspend fun runTask(name: String,): retrofit2.Response<MessageResponse>
    suspend fun updateTaskInterval(name: String, request: UpdateTaskIntervalRequest,): retrofit2.Response<MessageResponse>
    suspend fun getWebhooks(): retrofit2.Response<DataWrapper<List<WebhookDto>>>
    suspend fun createWebhook(request: CreateWebhookRequest,): retrofit2.Response<DataWrapper<WebhookDto>>
    suspend fun updateWebhook(webhookId: Int, request: UpdateWebhookRequest,): retrofit2.Response<DataWrapper<WebhookDto>>
    suspend fun deleteWebhook(webhookId: Int,): retrofit2.Response<MessageResponse>
    suspend fun testWebhook(webhookId: Int,): retrofit2.Response<DataWrapper<WebhookTestResponse>>
    suspend fun getAdminSettings(): retrofit2.Response<DataWrapper<AdminSettingsDto>>
    suspend fun updateAdminSettings(request: UpdateAdminSettingsRequest,): retrofit2.Response<DataWrapper<AdminSettingsDto>>
    suspend fun getProviderSettings(provider: String): retrofit2.Response<DataWrapper<ProviderSettingsDto>>
    suspend fun updateProviderSettings(provider: String, request: UpdateProviderRequest,): retrofit2.Response<DataWrapper<ProviderSettingsDto>>
    suspend fun getOpenSubsSettings(): retrofit2.Response<DataWrapper<OpenSubsSettingsDto>>
    suspend fun updateOpenSubsSettings(request: UpdateOpenSubsRequest,): retrofit2.Response<DataWrapper<OpenSubsSettingsDto>>
    suspend fun getAITranslationSettings(): retrofit2.Response<DataWrapper<AITranslationSettingsDto>>
    suspend fun updateAITranslationSettings(request: UpdateAITranslationRequest,): retrofit2.Response<DataWrapper<AITranslationSettingsDto>>
    suspend fun getAutoSubSettings(): retrofit2.Response<DataWrapper<AutoSubSettingsDto>>
    suspend fun updateAutoSubSettings(request: UpdateAutoSubRequest,): retrofit2.Response<DataWrapper<AutoSubSettingsDto>>
    suspend fun getAutoTranslateSettings(): retrofit2.Response<DataWrapper<AutoTranslateSettingsDto>>
    suspend fun updateAutoTranslateSettings(request: UpdateAutoTranslateRequest,): retrofit2.Response<DataWrapper<AutoTranslateSettingsDto>>
    suspend fun getPlaybackSettings(): retrofit2.Response<DataWrapper<PlaybackSettingsDto>>
    suspend fun updatePlaybackSettings(request: UpdatePlaybackRequest,): retrofit2.Response<DataWrapper<PlaybackSettingsDto>>
    suspend fun getCinemaSettings(): retrofit2.Response<DataWrapper<CinemaSettingsDto>>
    suspend fun updateCinemaSettings(request: UpdateCinemaRequest,): retrofit2.Response<DataWrapper<CinemaSettingsDto>>
    suspend fun getPretranscodeSettings(): retrofit2.Response<DataWrapper<PretranscodeSettingsDto>>
    suspend fun updatePretranscodeSettings(request: UpdatePretranscodeRequest,): retrofit2.Response<DataWrapper<PretranscodeSettingsDto>>
    suspend fun getMarkerStats(): retrofit2.Response<DataWrapper<MarkerStatsDto>>
    suspend fun startMarkerBackfill(request: BackfillMarkersRequest,): retrofit2.Response<DataWrapper<BackfillResultDto>>
    suspend fun getPretranscodeStatus(): retrofit2.Response<DataWrapper<PretranscodeStatusDto>>
    suspend fun getPretranscodeProfiles(): retrofit2.Response<DataWrapper<List<PretranscodeProfileDto>>>
    suspend fun togglePretranscodeProfile(id: Int, request: Map<String, Boolean>,): retrofit2.Response<DataWrapper<Any>>
    suspend fun getPretranscodeEstimate(libraryId: Int,): retrofit2.Response<DataWrapper<StorageEstimateDto>>
    suspend fun startPretranscode(): retrofit2.Response<DataWrapper<Any>>
    suspend fun stopPretranscode(): retrofit2.Response<DataWrapper<Any>>
    suspend fun resumePretranscode(): retrofit2.Response<DataWrapper<Any>>
    suspend fun cleanupPretranscodeFiles(): retrofit2.Response<DataWrapper<Any>>
    suspend fun refreshAllMetadata(): retrofit2.Response<DataWrapper<Map<String, Int>>>
    suspend fun searchSubtitles(mediaId: Int, lang: String,): retrofit2.Response<DataWrapper<List<SubtitleSearchResultDto>>>
    suspend fun downloadSubtitle(mediaId: Int, request: SubtitleDownloadRequest,): retrofit2.Response<DataWrapper<SubtitleDto>>
    suspend fun translateSubtitle(subtitleId: Int, request: TranslateSubtitleRequest,): retrofit2.Response<DataWrapper<SubtitleDto>>

    // ──────────────────────────────────────────────────────────
    // Migrated methods returning DataResult<T> (domain types)
    // New code should use these. Old Retrofit-typed methods above
    // will be removed once all consumers are migrated.
    // ──────────────────────────────────────────────────────────

    /** Profile */
    suspend fun fetchProfile(): com.velox.app.domain.model.DataResult<UserDto>
    suspend fun saveProfile(updates: Map<String, String>): com.velox.app.domain.model.DataResult<UserDto>

    /** Preferences */
    suspend fun fetchPreferences(): com.velox.app.domain.model.DataResult<UserPreferencesDto>
    suspend fun savePreferences(request: UpdatePreferencesRequest): com.velox.app.domain.model.DataResult<UserPreferencesDto>

    /** Sessions */
    suspend fun fetchSessions(): com.velox.app.domain.model.DataResult<List<SessionDto>>
    suspend fun revokeSessionSafe(sessionId: Int): com.velox.app.domain.model.DataResult<Unit>

    /** Security */
    suspend fun changePasswordSafe(request: ChangePasswordRequest): com.velox.app.domain.model.DataResult<Unit>

    /** Admin Settings */
    suspend fun fetchAdminSettings(): com.velox.app.domain.model.DataResult<AdminSettingsDto>
    suspend fun saveAdminSettings(request: UpdateAdminSettingsRequest): com.velox.app.domain.model.DataResult<AdminSettingsDto>

    /** Dashboard */
    suspend fun fetchServerInfo(): com.velox.app.domain.model.DataResult<ServerInfoDto>
    suspend fun fetchLibraryStats(): com.velox.app.domain.model.DataResult<List<LibraryStatsDto>>
    suspend fun fetchAdminActivity(limit: Int = 50, offset: Int = 0): com.velox.app.domain.model.DataResult<List<ActivityLogDto>>

    /** Admin Users */
    suspend fun fetchAdminUsers(): com.velox.app.domain.model.DataResult<List<AdminUserDto>>
    suspend fun createUserSafe(request: CreateUserRequest): com.velox.app.domain.model.DataResult<AdminUserDto>
    suspend fun updateUserSafe(userId: Int, request: UpdateUserRequest): com.velox.app.domain.model.DataResult<AdminUserDto>
    suspend fun deleteUserSafe(userId: Int): com.velox.app.domain.model.DataResult<Unit>

    /** Libraries */
    suspend fun createLibrarySafe(request: CreateLibraryRequest): com.velox.app.domain.model.DataResult<LibraryDto>
    suspend fun updateLibrarySafe(libraryId: Int, request: UpdateLibraryRequest): com.velox.app.domain.model.DataResult<LibraryDto>
    suspend fun deleteLibrarySafe(libraryId: Int): com.velox.app.domain.model.DataResult<Unit>
    suspend fun scanLibrarySafe(libraryId: Int, force: Boolean? = null): com.velox.app.domain.model.DataResult<Unit>

    /** Tasks */
    suspend fun fetchTasks(): com.velox.app.domain.model.DataResult<List<TaskDto>>
    suspend fun runTaskSafe(name: String): com.velox.app.domain.model.DataResult<Unit>
    suspend fun updateTaskIntervalSafe(name: String, request: UpdateTaskIntervalRequest): com.velox.app.domain.model.DataResult<Unit>

    /** Webhooks */
    suspend fun fetchWebhooks(): com.velox.app.domain.model.DataResult<List<WebhookDto>>
    suspend fun createWebhookSafe(request: CreateWebhookRequest): com.velox.app.domain.model.DataResult<WebhookDto>
    suspend fun updateWebhookSafe(webhookId: Int, request: UpdateWebhookRequest): com.velox.app.domain.model.DataResult<WebhookDto>
    suspend fun deleteWebhookSafe(webhookId: Int): com.velox.app.domain.model.DataResult<Unit>

    /** Admin - Playback/Cinema/Subtitles/Pretranscode/Markers */
    suspend fun fetchPlaybackSettings(): com.velox.app.domain.model.DataResult<PlaybackSettingsDto>
    suspend fun savePlaybackSettings(request: UpdatePlaybackRequest): com.velox.app.domain.model.DataResult<PlaybackSettingsDto>
    suspend fun fetchCinemaSettings(): com.velox.app.domain.model.DataResult<CinemaSettingsDto>
    suspend fun saveCinemaSettings(request: UpdateCinemaRequest): com.velox.app.domain.model.DataResult<CinemaSettingsDto>
    suspend fun fetchPretranscodeSettings(): com.velox.app.domain.model.DataResult<PretranscodeSettingsDto>
    suspend fun savePretranscodeSettings(request: UpdatePretranscodeRequest): com.velox.app.domain.model.DataResult<PretranscodeSettingsDto>
    suspend fun fetchMarkerStats(): com.velox.app.domain.model.DataResult<MarkerStatsDto>

    /** Subtitle Settings */
    suspend fun fetchSubtitleSettings(): com.velox.app.domain.model.DataResult<SubtitleSettingsBundle>

    /** Notifications */
    suspend fun fetchNotifications(limit: Int = 50, offset: Int = 0): com.velox.app.domain.model.DataResult<NotificationListResponse>
    suspend fun markNotificationsReadSafe(request: MarkReadRequest): com.velox.app.domain.model.DataResult<Unit>
    suspend fun markAllNotificationsReadSafe(): com.velox.app.domain.model.DataResult<Unit>
    suspend fun deleteNotificationsSafe(request: DeleteNotificationsRequest): com.velox.app.domain.model.DataResult<Unit>

    /** Provider Settings */
    suspend fun updateProviderSettingsSafe(provider: String, request: UpdateProviderRequest): com.velox.app.domain.model.DataResult<ProviderSettingsDto>
    suspend fun updateOpenSubsSettingsSafe(request: UpdateOpenSubsRequest): com.velox.app.domain.model.DataResult<OpenSubsSettingsDto>
    suspend fun updateAITranslationSettingsSafe(request: UpdateAITranslationRequest): com.velox.app.domain.model.DataResult<AITranslationSettingsDto>
    suspend fun updateAutoSubSettingsSafe(request: UpdateAutoSubRequest): com.velox.app.domain.model.DataResult<AutoSubSettingsDto>
    suspend fun fetchOpenSubsSettings(): com.velox.app.domain.model.DataResult<OpenSubsSettingsDto>
    suspend fun fetchAITranslationSettings(): com.velox.app.domain.model.DataResult<AITranslationSettingsDto>
    suspend fun fetchAutoSubSettings(): com.velox.app.domain.model.DataResult<AutoSubSettingsDto>
    suspend fun fetchAutoTranslateSettings(): com.velox.app.domain.model.DataResult<AutoTranslateSettingsDto>
    suspend fun updateAutoTranslateSettingsSafe(request: UpdateAutoTranslateRequest): com.velox.app.domain.model.DataResult<AutoTranslateSettingsDto>

    /** Pretranscode Actions */
    suspend fun fetchPretranscodeEstimate(libraryId: Int): com.velox.app.domain.model.DataResult<StorageEstimateDto>
    suspend fun fetchPretranscodeProfiles(): com.velox.app.domain.model.DataResult<List<PretranscodeProfileDto>>
    suspend fun fetchPretranscodeStatus(): com.velox.app.domain.model.DataResult<PretranscodeStatusDto>
    suspend fun togglePretranscodeProfileSafe(id: Int, request: Map<String, Boolean>): com.velox.app.domain.model.DataResult<Unit>
    suspend fun executePretranscodeActionSafe(action: String): com.velox.app.domain.model.DataResult<Unit>

    /** Markers */
    suspend fun startMarkerBackfillSafe(request: BackfillMarkersRequest): com.velox.app.domain.model.DataResult<BackfillResultDto>

    /** Provider Settings (read) */
    suspend fun fetchProviderSettings(provider: String): com.velox.app.domain.model.DataResult<ProviderSettingsDto>

    /** App Version */
    suspend fun fetchLatestAppVersion(platform: String = "android"): com.velox.app.domain.model.DataResult<AppVersionDto>

    /** Refresh All Metadata */
    suspend fun refreshAllMetadataSafe(): com.velox.app.domain.model.DataResult<Unit>
}

/**
 * Bundle of all subtitle-related settings fetched together.
 */
data class SubtitleSettingsBundle(
    val provider: ProviderSettingsDto?,
    val openSubs: OpenSubsSettingsDto?,
    val aiTranslation: AITranslationSettingsDto?,
    val autoSub: AutoSubSettingsDto?,
    val autoTranslate: AutoTranslateSettingsDto?,
)
