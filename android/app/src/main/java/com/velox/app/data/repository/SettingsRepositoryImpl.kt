package com.velox.app.data.repository

import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.*
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.repository.SettingsRepository
import retrofit2.Response
import javax.inject.Inject

class SettingsRepositoryImpl @Inject constructor(
    private val api: VeloxApiProvider
) : SettingsRepository {

    override suspend fun healthCheck(): retrofit2.Response<MessageResponse> {
        return api.getApi().healthCheck()
    }
    override suspend fun refreshToken(request: RefreshRequest): retrofit2.Response<DataWrapper<RefreshResponse>> {
        return api.getApi().refreshToken(request)
    }
    override suspend fun logout(): retrofit2.Response<MessageResponse> {
        return api.getApi().logout()
    }
    override suspend fun changePassword(request: ChangePasswordRequest): retrofit2.Response<MessageResponse> {
        return api.getApi().changePassword(request)
    }
    override suspend fun getMe(): retrofit2.Response<DataWrapper<UserDto>> {
        return api.getApi().getMe()
    }
    override suspend fun getSessions(): retrofit2.Response<DataWrapper<List<SessionDto>>> {
        return api.getApi().getSessions()
    }
    override suspend fun revokeSession(sessionId: Int): retrofit2.Response<MessageResponse> {
        return api.getApi().revokeSession(sessionId)
    }
    override suspend fun getMediaList(type: String?, genre: String?, year: String?, minRating: Float?, sort: String?, search: String?, libraryId: Int?, limit: Int, offset: Int): retrofit2.Response<DataWrapper<List<MediaListItemDto>>> {
        return api.getApi().getMediaList(type, genre, year, minRating, sort, search, libraryId, limit, offset)
    }
    override suspend fun getMedia(mediaId: Int): retrofit2.Response<DataWrapper<MediaDto>> {
        return api.getApi().getMedia(mediaId)
    }
    override suspend fun getMediaWithFiles(mediaId: Int): retrofit2.Response<DataWrapper<MediaWithFilesDto>> {
        return api.getApi().getMediaWithFiles(mediaId)
    }
    override suspend fun getMediaCredits(mediaId: Int): retrofit2.Response<DataWrapper<CreditsDto>> {
        return api.getApi().getMediaCredits(mediaId)
    }
    override suspend fun getSeriesList(genre: String?, year: String?, sort: String?, search: String?, libraryId: Int?, limit: Int, offset: Int): retrofit2.Response<DataWrapper<List<SeriesListItemDto>>> {
        return api.getApi().getSeriesList(genre, year, sort, search, libraryId, limit, offset)
    }
    override suspend fun getSeriesCredits(seriesId: Int): retrofit2.Response<DataWrapper<CreditsDto>> {
        return api.getApi().getSeriesCredits(seriesId)
    }
    override suspend fun browse(libraryId: Int?, path: String?): retrofit2.Response<DataWrapper<BrowseResponse>> {
        return api.getApi().browse(libraryId, path)
    }
    override suspend fun getGenres(type: String?): retrofit2.Response<DataWrapper<List<GenreDto>>> {
        return api.getApi().getGenres(type)
    }
    override suspend fun search(query: String, limit: Int): retrofit2.Response<DataWrapper<SearchResponse>> {
        return api.getApi().search(query, limit)
    }
    override suspend fun getPlaybackInfo(mediaId: Int, request: PlaybackInfoRequest?): retrofit2.Response<DataWrapper<PlaybackInfoDto>> {
        return api.getApi().getPlaybackInfo(mediaId, request)
    }
    override suspend fun getStreamUrl(mediaId: Int): retrofit2.Response<DataWrapper<StreamUrlResponse>> {
        return api.getApi().getStreamUrl(mediaId)
    }
    override suspend fun getProfile(): retrofit2.Response<DataWrapper<UserDto>> {
        return api.getApi().getProfile()
    }
    override suspend fun updateProfile(updates: Map<String, String>): retrofit2.Response<DataWrapper<UserDto>> {
        return api.getApi().updateProfile(updates)
    }
    override suspend fun getPreferences(): retrofit2.Response<DataWrapper<UserPreferencesDto>> {
        return api.getApi().getPreferences()
    }
    override suspend fun updatePreferences(request: UpdatePreferencesRequest): retrofit2.Response<DataWrapper<UserPreferencesDto>> {
        return api.getApi().updatePreferences(request)
    }
    override suspend fun getProgress(mediaId: Int): retrofit2.Response<DataWrapper<UserDataDto?>> {
        return api.getApi().getProgress(mediaId)
    }
    override suspend fun updateProgress(mediaId: Int, request: UpdateProgressRequest,): retrofit2.Response<MessageResponse> {
        return api.getApi().updateProgress(mediaId, request)
    }
    override suspend fun dismissProgress(mediaId: Int): retrofit2.Response<MessageResponse> {
        return api.getApi().dismissProgress(mediaId)
    }
    override suspend fun getFavorites(limit: Int, offset: Int): retrofit2.Response<DataWrapper<List<MediaListItemDto>>> {
        return api.getApi().getFavorites(limit, offset)
    }
    override suspend fun toggleFavorite(mediaId: Int): retrofit2.Response<DataWrapper<ToggleFavoriteResponse>> {
        return api.getApi().toggleFavorite(mediaId)
    }
    override suspend fun getRecentlyWatched(limit: Int): retrofit2.Response<DataWrapper<List<MediaListItemDto>>> {
        return api.getApi().getRecentlyWatched(limit)
    }
    override suspend fun getContinueWatching(limit: Int): retrofit2.Response<DataWrapper<List<ContinueWatchingItemDto>>> {
        return api.getApi().getContinueWatching(limit)
    }
    override suspend fun getNextUp(limit: Int): retrofit2.Response<DataWrapper<List<NextUpItemDto>>> {
        return api.getApi().getNextUp(limit)
    }
    override suspend fun getNotifications(limit: Int, offset: Int): retrofit2.Response<DataWrapper<NotificationListResponse>> {
        return api.getApi().getNotifications(limit, offset)
    }
    override suspend fun markNotificationsRead(request: MarkReadRequest,): retrofit2.Response<MessageResponse> {
        return api.getApi().markNotificationsRead(request)
    }
    override suspend fun markAllNotificationsRead(): retrofit2.Response<MessageResponse> {
        return api.getApi().markAllNotificationsRead()
    }
    override suspend fun deleteNotifications(request: DeleteNotificationsRequest,): retrofit2.Response<MessageResponse> {
        return api.getApi().deleteNotifications(request)
    }
    override suspend fun getUnreadCount(): retrofit2.Response<DataWrapper<UnreadCountResponse>> {
        return api.getApi().getUnreadCount()
    }
    override suspend fun getMediaGenres(mediaId: Int,): retrofit2.Response<DataWrapper<List<GenreDto>>> {
        return api.getApi().getMediaGenres(mediaId)
    }
    override suspend fun getMediaCinema(mediaId: Int,): retrofit2.Response<DataWrapper<CinemaDto>> {
        return api.getApi().getMediaCinema(mediaId)
    }
    override suspend fun refreshMediaMetadata(mediaId: Int,): retrofit2.Response<MessageResponse> {
        return api.getApi().refreshMediaMetadata(mediaId)
    }
    override suspend fun updateMediaMetadata(mediaId: Int, request: UpdateMetadataRequest,): retrofit2.Response<DataWrapper<MediaDto>> {
        return api.getApi().updateMediaMetadata(mediaId, request)
    }
    override suspend fun updateMediaImages(mediaId: Int, request: UpdateImagesRequest,): retrofit2.Response<MessageResponse> {
        return api.getApi().updateMediaImages(mediaId, request)
    }
    override suspend fun getSeriesGenres(seriesId: Int,): retrofit2.Response<DataWrapper<List<GenreDto>>> {
        return api.getApi().getSeriesGenres(seriesId)
    }
    override suspend fun getSeriesCinema(seriesId: Int,): retrofit2.Response<DataWrapper<CinemaDto>> {
        return api.getApi().getSeriesCinema(seriesId)
    }
    override suspend fun getServerInfo(): retrofit2.Response<DataWrapper<ServerInfoDto>> {
        return api.getApi().getServerInfo()
    }
    override suspend fun getLatestAppVersion(platform: String): retrofit2.Response<AppVersionDto> {
        return api.getApi().getLatestAppVersion(platform)
    }
    override suspend fun getLibraryStats(): retrofit2.Response<DataWrapper<List<LibraryStatsDto>>> {
        return api.getApi().getLibraryStats()
    }
    override suspend fun createLibrary(request: CreateLibraryRequest,): retrofit2.Response<DataWrapper<LibraryDto>> {
        return api.getApi().createLibrary(request)
    }
    override suspend fun updateLibrary(libraryId: Int, request: UpdateLibraryRequest,): retrofit2.Response<DataWrapper<LibraryDto>> {
        return api.getApi().updateLibrary(libraryId, request)
    }
    override suspend fun deleteLibrary(libraryId: Int,): retrofit2.Response<MessageResponse> {
        return api.getApi().deleteLibrary(libraryId)
    }
    override suspend fun scanLibrary(libraryId: Int, force: Boolean?): retrofit2.Response<MessageResponse> {
        return api.getApi().scanLibrary(libraryId, force)
    }
    override suspend fun getAdminUsers(): retrofit2.Response<DataWrapper<List<AdminUserDto>>> {
        return api.getApi().getAdminUsers()
    }
    override suspend fun createUser(request: CreateUserRequest,): retrofit2.Response<DataWrapper<AdminUserDto>> {
        return api.getApi().createUser(request)
    }
    override suspend fun updateUser(userId: Int, request: UpdateUserRequest,): retrofit2.Response<DataWrapper<AdminUserDto>> {
        return api.getApi().updateUser(userId, request)
    }
    override suspend fun deleteUser(userId: Int,): retrofit2.Response<MessageResponse> {
        return api.getApi().deleteUser(userId)
    }
    override suspend fun getAdminActivity(limit: Int, offset: Int): retrofit2.Response<DataWrapper<List<ActivityLogDto>>> {
        return api.getApi().getAdminActivity(limit, offset)
    }
    override suspend fun getTasks(): retrofit2.Response<DataWrapper<List<TaskDto>>> {
        return api.getApi().getTasks()
    }
    override suspend fun runTask(name: String,): retrofit2.Response<MessageResponse> {
        return api.getApi().runTask(name)
    }
    override suspend fun updateTaskInterval(name: String, request: UpdateTaskIntervalRequest,): retrofit2.Response<MessageResponse> {
        return api.getApi().updateTaskInterval(name, request)
    }
    override suspend fun getWebhooks(): retrofit2.Response<DataWrapper<List<WebhookDto>>> {
        return api.getApi().getWebhooks()
    }
    override suspend fun createWebhook(request: CreateWebhookRequest,): retrofit2.Response<DataWrapper<WebhookDto>> {
        return api.getApi().createWebhook(request)
    }
    override suspend fun updateWebhook(webhookId: Int, request: UpdateWebhookRequest,): retrofit2.Response<DataWrapper<WebhookDto>> {
        return api.getApi().updateWebhook(webhookId, request)
    }
    override suspend fun deleteWebhook(webhookId: Int,): retrofit2.Response<MessageResponse> {
        return api.getApi().deleteWebhook(webhookId)
    }
    override suspend fun testWebhook(webhookId: Int,): retrofit2.Response<DataWrapper<WebhookTestResponse>> {
        return api.getApi().testWebhook(webhookId)
    }
    override suspend fun getAdminSettings(): retrofit2.Response<DataWrapper<AdminSettingsDto>> {
        return api.getApi().getAdminSettings()
    }
    override suspend fun updateAdminSettings(request: UpdateAdminSettingsRequest,): retrofit2.Response<DataWrapper<AdminSettingsDto>> {
        return api.getApi().updateAdminSettings(request)
    }
    override suspend fun getProviderSettings(provider: String): retrofit2.Response<DataWrapper<ProviderSettingsDto>> {
        return api.getApi().getProviderSettings(provider)
    }
    override suspend fun updateProviderSettings(provider: String, request: UpdateProviderRequest,): retrofit2.Response<DataWrapper<ProviderSettingsDto>> {
        return api.getApi().updateProviderSettings(provider, request)
    }
    override suspend fun getOpenSubsSettings(): retrofit2.Response<DataWrapper<OpenSubsSettingsDto>> {
        return api.getApi().getOpenSubsSettings()
    }
    override suspend fun updateOpenSubsSettings(request: UpdateOpenSubsRequest,): retrofit2.Response<DataWrapper<OpenSubsSettingsDto>> {
        return api.getApi().updateOpenSubsSettings(request)
    }
    override suspend fun getAITranslationSettings(): retrofit2.Response<DataWrapper<AITranslationSettingsDto>> {
        return api.getApi().getAITranslationSettings()
    }
    override suspend fun updateAITranslationSettings(request: UpdateAITranslationRequest,): retrofit2.Response<DataWrapper<AITranslationSettingsDto>> {
        return api.getApi().updateAITranslationSettings(request)
    }
    override suspend fun getAutoSubSettings(): retrofit2.Response<DataWrapper<AutoSubSettingsDto>> {
        return api.getApi().getAutoSubSettings()
    }
    override suspend fun updateAutoSubSettings(request: UpdateAutoSubRequest,): retrofit2.Response<DataWrapper<AutoSubSettingsDto>> {
        return api.getApi().updateAutoSubSettings(request)
    }
    override suspend fun getAutoTranslateSettings(): retrofit2.Response<DataWrapper<AutoTranslateSettingsDto>> {
        return api.getApi().getAutoTranslateSettings()
    }
    override suspend fun updateAutoTranslateSettings(request: UpdateAutoTranslateRequest,): retrofit2.Response<DataWrapper<AutoTranslateSettingsDto>> {
        return api.getApi().updateAutoTranslateSettings(request)
    }
    override suspend fun getPlaybackSettings(): retrofit2.Response<DataWrapper<PlaybackSettingsDto>> {
        return api.getApi().getPlaybackSettings()
    }
    override suspend fun updatePlaybackSettings(request: UpdatePlaybackRequest,): retrofit2.Response<DataWrapper<PlaybackSettingsDto>> {
        return api.getApi().updatePlaybackSettings(request)
    }
    override suspend fun getCinemaSettings(): retrofit2.Response<DataWrapper<CinemaSettingsDto>> {
        return api.getApi().getCinemaSettings()
    }
    override suspend fun updateCinemaSettings(request: UpdateCinemaRequest,): retrofit2.Response<DataWrapper<CinemaSettingsDto>> {
        return api.getApi().updateCinemaSettings(request)
    }
    override suspend fun getPretranscodeSettings(): retrofit2.Response<DataWrapper<PretranscodeSettingsDto>> {
        return api.getApi().getPretranscodeSettings()
    }
    override suspend fun updatePretranscodeSettings(request: UpdatePretranscodeRequest,): retrofit2.Response<DataWrapper<PretranscodeSettingsDto>> {
        return api.getApi().updatePretranscodeSettings(request)
    }
    override suspend fun getMarkerStats(): retrofit2.Response<DataWrapper<MarkerStatsDto>> {
        return api.getApi().getMarkerStats()
    }
    override suspend fun startMarkerBackfill(request: BackfillMarkersRequest,): retrofit2.Response<DataWrapper<BackfillResultDto>> {
        return api.getApi().startMarkerBackfill(request)
    }
    override suspend fun getPretranscodeStatus(): retrofit2.Response<DataWrapper<PretranscodeStatusDto>> {
        return api.getApi().getPretranscodeStatus()
    }
    override suspend fun getPretranscodeProfiles(): retrofit2.Response<DataWrapper<List<PretranscodeProfileDto>>> {
        return api.getApi().getPretranscodeProfiles()
    }
    override suspend fun togglePretranscodeProfile(id: Int, request: Map<String, Boolean>,): retrofit2.Response<DataWrapper<Any>> {
        return api.getApi().togglePretranscodeProfile(id, request)
    }
    override suspend fun getPretranscodeEstimate(libraryId: Int,): retrofit2.Response<DataWrapper<StorageEstimateDto>> {
        return api.getApi().getPretranscodeEstimate(libraryId)
    }
    override suspend fun startPretranscode(): retrofit2.Response<DataWrapper<Any>> {
        return api.getApi().startPretranscode()
    }
    override suspend fun stopPretranscode(): retrofit2.Response<DataWrapper<Any>> {
        return api.getApi().stopPretranscode()
    }
    override suspend fun resumePretranscode(): retrofit2.Response<DataWrapper<Any>> {
        return api.getApi().resumePretranscode()
    }
    override suspend fun cleanupPretranscodeFiles(): retrofit2.Response<DataWrapper<Any>> {
        return api.getApi().cleanupPretranscodeFiles()
    }
    override suspend fun refreshAllMetadata(): retrofit2.Response<DataWrapper<Map<String, Int>>> {
        return api.getApi().refreshAllMetadata()
    }
    override suspend fun searchSubtitles(mediaId: Int, lang: String,): retrofit2.Response<DataWrapper<List<SubtitleSearchResultDto>>> {
        return api.getApi().searchSubtitles(mediaId, lang)
    }
    override suspend fun downloadSubtitle(mediaId: Int, request: SubtitleDownloadRequest,): retrofit2.Response<DataWrapper<SubtitleDto>> {
        return api.getApi().downloadSubtitle(mediaId, request)
    }
    override suspend fun translateSubtitle(subtitleId: Int, request: TranslateSubtitleRequest,): retrofit2.Response<DataWrapper<SubtitleDto>> {
        return api.getApi().translateSubtitle(subtitleId, request)
    }

    // ──────────────────────────────────────────────────────────
    // Migrated methods returning DataResult<T>
    // ──────────────────────────────────────────────────────────

    override suspend fun fetchProfile() =
        safeApiCall({ api.getApi().getProfile() }) { it }

    override suspend fun saveProfile(updates: Map<String, String>) =
        safeApiCall({ api.getApi().updateProfile(updates) }) { it }

    override suspend fun fetchPreferences() =
        safeApiCall({ api.getApi().getPreferences() }) { it }

    override suspend fun savePreferences(request: UpdatePreferencesRequest) =
        safeApiCall({ api.getApi().updatePreferences(request) }) { it }

    override suspend fun fetchSessions() =
        safeApiCall({ api.getApi().getSessions() }) { it }

    override suspend fun revokeSessionSafe(sessionId: Int) =
        safeMessageCall { api.getApi().revokeSession(sessionId) }

    override suspend fun changePasswordSafe(request: ChangePasswordRequest) =
        safeMessageCall { api.getApi().changePassword(request) }

    override suspend fun fetchAdminSettings() =
        safeApiCall({ api.getApi().getAdminSettings() }) { it }

    override suspend fun saveAdminSettings(request: UpdateAdminSettingsRequest) =
        safeApiCall({ api.getApi().updateAdminSettings(request) }) { it }

    override suspend fun fetchServerInfo() =
        safeApiCall({ api.getApi().getServerInfo() }) { it }

    override suspend fun fetchLibraryStats() =
        safeApiCall({ api.getApi().getLibraryStats() }) { it }

    override suspend fun fetchAdminActivity(limit: Int, offset: Int) =
        safeApiCall({ api.getApi().getAdminActivity(limit, offset) }) { it }

    override suspend fun fetchAdminUsers() =
        safeApiCall({ api.getApi().getAdminUsers() }) { it }

    override suspend fun createUserSafe(request: CreateUserRequest) =
        safeApiCall({ api.getApi().createUser(request) }) { it }

    override suspend fun updateUserSafe(userId: Int, request: UpdateUserRequest) =
        safeApiCall({ api.getApi().updateUser(userId, request) }) { it }

    override suspend fun deleteUserSafe(userId: Int) =
        safeMessageCall { api.getApi().deleteUser(userId) }

    override suspend fun createLibrarySafe(request: CreateLibraryRequest) =
        safeApiCall({ api.getApi().createLibrary(request) }) { it }

    override suspend fun updateLibrarySafe(libraryId: Int, request: UpdateLibraryRequest) =
        safeApiCall({ api.getApi().updateLibrary(libraryId, request) }) { it }

    override suspend fun deleteLibrarySafe(libraryId: Int) =
        safeMessageCall { api.getApi().deleteLibrary(libraryId) }

    override suspend fun scanLibrarySafe(libraryId: Int, force: Boolean?) =
        safeMessageCall { api.getApi().scanLibrary(libraryId, force) }

    override suspend fun fetchTasks() =
        safeApiCall({ api.getApi().getTasks() }) { it }

    override suspend fun runTaskSafe(name: String) =
        safeMessageCall { api.getApi().runTask(name) }

    override suspend fun updateTaskIntervalSafe(name: String, request: UpdateTaskIntervalRequest) =
        safeMessageCall { api.getApi().updateTaskInterval(name, request) }

    override suspend fun fetchWebhooks() =
        safeApiCall({ api.getApi().getWebhooks() }) { it }

    override suspend fun createWebhookSafe(request: CreateWebhookRequest) =
        safeApiCall({ api.getApi().createWebhook(request) }) { it }

    override suspend fun updateWebhookSafe(webhookId: Int, request: UpdateWebhookRequest) =
        safeApiCall({ api.getApi().updateWebhook(webhookId, request) }) { it }

    override suspend fun deleteWebhookSafe(webhookId: Int) =
        safeMessageCall { api.getApi().deleteWebhook(webhookId) }

    override suspend fun fetchPlaybackSettings() =
        safeApiCall({ api.getApi().getPlaybackSettings() }) { it }

    override suspend fun savePlaybackSettings(request: UpdatePlaybackRequest) =
        safeApiCall({ api.getApi().updatePlaybackSettings(request) }) { it }

    override suspend fun fetchCinemaSettings() =
        safeApiCall({ api.getApi().getCinemaSettings() }) { it }

    override suspend fun saveCinemaSettings(request: UpdateCinemaRequest) =
        safeApiCall({ api.getApi().updateCinemaSettings(request) }) { it }

    override suspend fun fetchPretranscodeSettings() =
        safeApiCall({ api.getApi().getPretranscodeSettings() }) { it }

    override suspend fun savePretranscodeSettings(request: UpdatePretranscodeRequest) =
        safeApiCall({ api.getApi().updatePretranscodeSettings(request) }) { it }

    override suspend fun fetchMarkerStats() =
        safeApiCall({ api.getApi().getMarkerStats() }) { it }

    override suspend fun fetchSubtitleSettings(): com.velox.app.domain.model.DataResult<com.velox.app.domain.repository.SubtitleSettingsBundle> {
        return try {
            val provider = try { api.getApi().getProviderSettings("tmdb").body()?.data } catch (_: Exception) { null }
            val openSubs = try { api.getApi().getOpenSubsSettings().body()?.data } catch (_: Exception) { null }
            val aiTranslation = try { api.getApi().getAITranslationSettings().body()?.data } catch (_: Exception) { null }
            val autoSub = try { api.getApi().getAutoSubSettings().body()?.data } catch (_: Exception) { null }
            val autoTranslate = try { api.getApi().getAutoTranslateSettings().body()?.data } catch (_: Exception) { null }
            com.velox.app.domain.model.DataResult.success(
                com.velox.app.domain.repository.SubtitleSettingsBundle(provider, openSubs, aiTranslation, autoSub, autoTranslate)
            )
        } catch (e: Exception) {
            com.velox.app.domain.model.DataResult.unknown(e)
        }
    }

    // ── Notifications ──

    override suspend fun fetchNotifications(limit: Int, offset: Int) =
        safeApiCall({ api.getApi().getNotifications(limit, offset) }) { it }

    override suspend fun markNotificationsReadSafe(request: MarkReadRequest) =
        safeMessageCall { api.getApi().markNotificationsRead(request) }

    override suspend fun markAllNotificationsReadSafe() =
        safeMessageCall { api.getApi().markAllNotificationsRead() }

    override suspend fun deleteNotificationsSafe(request: DeleteNotificationsRequest) =
        safeMessageCall { api.getApi().deleteNotifications(request) }

    // ── Provider Settings ──

    override suspend fun updateProviderSettingsSafe(provider: String, request: UpdateProviderRequest) =
        safeApiCall({ api.getApi().updateProviderSettings(provider, request) }) { it }

    override suspend fun updateOpenSubsSettingsSafe(request: UpdateOpenSubsRequest) =
        safeApiCall({ api.getApi().updateOpenSubsSettings(request) }) { it }

    override suspend fun updateAITranslationSettingsSafe(request: UpdateAITranslationRequest) =
        safeApiCall({ api.getApi().updateAITranslationSettings(request) }) { it }

    override suspend fun updateAutoSubSettingsSafe(request: UpdateAutoSubRequest) =
        safeApiCall({ api.getApi().updateAutoSubSettings(request) }) { it }

    override suspend fun fetchOpenSubsSettings() =
        safeApiCall({ api.getApi().getOpenSubsSettings() }) { it }

    override suspend fun fetchAITranslationSettings() =
        safeApiCall({ api.getApi().getAITranslationSettings() }) { it }

    override suspend fun fetchAutoSubSettings() =
        safeApiCall({ api.getApi().getAutoSubSettings() }) { it }

    override suspend fun fetchAutoTranslateSettings() =
        safeApiCall({ api.getApi().getAutoTranslateSettings() }) { it }

    override suspend fun updateAutoTranslateSettingsSafe(request: UpdateAutoTranslateRequest) =
        safeApiCall({ api.getApi().updateAutoTranslateSettings(request) }) { it }

    // ── Pretranscode Actions ──

    override suspend fun fetchPretranscodeEstimate(libraryId: Int) =
        safeApiCall({ api.getApi().getPretranscodeEstimate(libraryId) }) { it }

    override suspend fun togglePretranscodeProfileSafe(id: Int, request: Map<String, Boolean>) =
        safeApiCall({ api.getApi().togglePretranscodeProfile(id, request) }) { }

    override suspend fun executePretranscodeActionSafe(action: String): com.velox.app.domain.model.DataResult<Unit> {
        val call: suspend () -> retrofit2.Response<DataWrapper<Any>> = when (action) {
            "start" -> { { api.getApi().startPretranscode() } }
            "stop" -> { { api.getApi().stopPretranscode() } }
            "resume" -> { { api.getApi().resumePretranscode() } }
            "cleanup" -> { { api.getApi().cleanupPretranscodeFiles() } }
            else -> return com.velox.app.domain.model.DataResult.error(
                com.velox.app.domain.model.DataError.Kind.Validation,
                "Unknown action: $action",
            )
        }
        return safeApiCall(call) { }
    }

    // ── Markers ──

    override suspend fun startMarkerBackfillSafe(request: BackfillMarkersRequest) =
        safeApiCall({ api.getApi().startMarkerBackfill(request) }) { it }

    // ── App Version ──

    override suspend fun fetchLatestAppVersion(platform: String) =
        safeRawApiCall({ api.getApi().getLatestAppVersion(platform) }) { it }

    override suspend fun fetchPretranscodeProfiles() =
        safeApiCall({ api.getApi().getPretranscodeProfiles() }) { it }

    override suspend fun fetchPretranscodeStatus() =
        safeApiCall({ api.getApi().getPretranscodeStatus() }) { it }

    override suspend fun fetchProviderSettings(provider: String) =
        safeApiCall({ api.getApi().getProviderSettings(provider) }) { it }

    override suspend fun refreshAllMetadataSafe(): DataResult<Unit> =
        safeApiCall({ api.getApi().refreshAllMetadata() }) { _ -> }
}
