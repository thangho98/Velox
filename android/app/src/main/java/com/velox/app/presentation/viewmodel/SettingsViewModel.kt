package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.VeloxApi
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.*
import com.velox.app.domain.model.User
import com.velox.app.domain.repository.AuthRepository
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class SettingsSection(val title: String, val adminOnly: Boolean = false) {
    PROFILE("Profile"),
    PREFERENCES("Preferences"),
    SECURITY("Security"),
    SESSIONS("Sessions"),
    METADATA("Metadata"),
    SUBTITLES("Subtitles"),
    PLAYBACK("Playback"),
    CINEMA("Cinema Mode"),
    PRETRANSCODE("Pre-transcode"),
    MARKERS("Skip Intro / Credits"),
    DASHBOARD("Dashboard", adminOnly = true),
    LIBRARIES("Libraries", adminOnly = true),
    USERS("Users", adminOnly = true),
    ACTIVITY("Activity", adminOnly = true),
    TASKS("Tasks", adminOnly = true),
    WEBHOOKS("Webhooks", adminOnly = true),
}

data class UserPreferences(
    val subtitleLanguage: String? = null,
    val audioLanguage: String? = null,
    val maxStreamingQuality: String? = null,
    val theme: String? = null,
    val language: String? = null,
)

data class Session(
    val id: Int,
    val deviceName: String,
    val ipAddress: String,
    val lastActive: String,
    val isCurrent: Boolean,
)

data class NotificationSettings(
    val newContent: Boolean = true,
    val libraryUpdates: Boolean = true,
    val systemNotifications: Boolean = true,
)

// Deleted PlaybackSettings dummy class

data class SubtitleSettings(
    val defaultLanguage: String = "English",
    val autoLoad: Boolean = true,
    val fontSize: String = "Medium",
    val background: Boolean = false,
)

// Deleted CinemaSettings mockup

data class SkipMarkerSettings(
    val skipIntro: Boolean = true,
    val skipCredits: Boolean = true,
    val autoSkipSponsor: Boolean = true,
)



data class AdminUser(
    val id: Int,
    val username: String,
    val displayName: String,
    val isAdmin: Boolean,
    val lastActive: String?,
    val createdAt: String,
)



data class Webhook(
    val id: Int,
    val url: String,
    val events: List<String>,
    val active: Boolean,
)

data class SettingsUiState(
    val isLoading: Boolean = false,
    val currentSection: SettingsSection = SettingsSection.PROFILE,
    val selectedTab: String = "menu", // Tab navigation: profile/preferences/security/sessions/cinema/subtitles/admin
    val user: User? = null,
    val isAdmin: Boolean = false,

    // Profile
    val displayName: String = "",
    val username: String = "",

    // Preferences
    val preferences: UserPreferences = UserPreferences(),

    // Security
    val currentPassword: String = "",
    val newPassword: String = "",
    val confirmPassword: String = "",
    val securityError: String? = null,
    val securitySuccess: String? = null,

    // Sessions
    val sessions: List<Session> = emptyList(),
    val sessionRevokeError: String? = null,

    // Notifications
    val notificationsList: List<com.velox.app.data.model.NotificationDto> = emptyList(),
    val unreadNotificationsCount: Int = 0,

    // Metadata
    val tmdbSettings: com.velox.app.data.model.ProviderSettingsDto? = null,
    val omdbSettings: com.velox.app.data.model.ProviderSettingsDto? = null,
    val tvdbSettings: com.velox.app.data.model.ProviderSettingsDto? = null,
    val fanartSettings: com.velox.app.data.model.ProviderSettingsDto? = null,

    // Admin Subtitles
    val openSubsSettings: com.velox.app.data.model.OpenSubsSettingsDto? = null,
    val subdlSettings: com.velox.app.data.model.ProviderSettingsDto? = null,
    val deepLSettings: com.velox.app.data.model.ProviderSettingsDto? = null,
    val aiTranslationSettings: com.velox.app.data.model.AITranslationSettingsDto? = null,
    val autoSubSettings: com.velox.app.data.model.AutoSubSettingsDto? = null,

    // Admin Playback
    val adminPlaybackSettings: com.velox.app.data.model.PlaybackSettingsDto? = null,

    // Admin Cinema
    val adminCinemaSettings: com.velox.app.data.model.CinemaSettingsDto? = null,

    // Admin Pre-transcode
    val pretranscodeSettings: com.velox.app.data.model.PretranscodeSettingsDto? = null,
    val pretranscodeStatus: com.velox.app.data.model.PretranscodeStatusDto? = null,
    val pretranscodeProfiles: List<com.velox.app.data.model.PretranscodeProfileDto> = emptyList(),
    val pretranscodeEstimate: com.velox.app.data.model.StorageEstimateDto? = null,

    // Admin - Markers Dashboard
    val markerStats: com.velox.app.data.model.MarkerStatsDto? = null,
    val isRunningDetection: Boolean = false,
    
    // Skip Markers (Preferences)
    val skipMarkers: SkipMarkerSettings = SkipMarkerSettings(),

    // Admin - Dashboard
    val serverInfo: com.velox.app.data.model.ServerInfoDto? = null,
    val libraryStats: List<com.velox.app.data.model.LibraryStatsDto> = emptyList(),
    val activityLogs: List<com.velox.app.data.model.ActivityLogDto> = emptyList(),

    // Admin - Libraries
    val libraries: List<com.velox.app.domain.model.Library> = emptyList(),
    val scanningLibraryId: Int? = null,

    // Admin - Users
    val users: List<AdminUser> = emptyList(),

    // Admin - Tasks
    val tasks: List<com.velox.app.data.model.TaskDto> = emptyList(),
    val runningTaskName: String? = null,

    // Admin - Webhooks
    val webhooks: List<Webhook> = emptyList(),

    // App Version
    val latestVersion: String? = null,
    val updateDownloadUrl: String? = null,
    val updateIsMandatory: Boolean = false,
    val updateStatus: String = "",
    val isCheckingUpdate: Boolean = false,

    // General
    val error: String? = null,
    val successMessage: String? = null,
)

@HiltViewModel
class SettingsViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val mediaRepository: MediaRepository,
    private val apiProvider: VeloxApiProvider,
) : ViewModel() {

    private val veloxApi: VeloxApi
        get() = apiProvider.getApi()

    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    init {
        loadUserInfo()
    }

    private fun loadUserInfo() {
        viewModelScope.launch {
            authRepository.getCurrentUser().collect { user ->
                user?.let {
                    _uiState.update { state ->
                        state.copy(
                            user = it,
                            isAdmin = it.isAdmin,
                            displayName = it.displayName,
                            username = it.username,
                        )
                    }
                    // Load initial section data
                    loadSectionData(_uiState.value.currentSection)
                }
            }
        }
    }

    fun selectSection(section: SettingsSection) {
        // Check admin permission
        if (section.adminOnly && !(_uiState.value.isAdmin)) {
            return
        }
        _uiState.update { it.copy(currentSection = section, error = null, successMessage = null) }
        loadSectionData(section)
    }

    fun setSelectedTab(tab: String) {
        _uiState.update { it.copy(selectedTab = tab, error = null, successMessage = null) }
        // Map tab to corresponding section and load data
        val section = when (tab) {
            "profile" -> SettingsSection.PROFILE
            "preferences" -> SettingsSection.PREFERENCES
            "security" -> SettingsSection.SECURITY
            "sessions" -> SettingsSection.SESSIONS
            "metadata" -> SettingsSection.METADATA
            "subtitles" -> SettingsSection.SUBTITLES
            "playback" -> SettingsSection.PLAYBACK
            "cinema" -> SettingsSection.CINEMA
            "pretranscode" -> SettingsSection.PRETRANSCODE
            "markers" -> SettingsSection.MARKERS
            "dashboard" -> SettingsSection.DASHBOARD
            "libraries" -> SettingsSection.LIBRARIES
            "users" -> SettingsSection.USERS
            "activity" -> SettingsSection.ACTIVITY
            "tasks" -> SettingsSection.TASKS
            "webhooks" -> SettingsSection.WEBHOOKS
            else -> SettingsSection.PROFILE
        }
        _uiState.update { it.copy(currentSection = section) }
        loadSectionData(section)
    }

    private fun loadSectionData(section: SettingsSection) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            when (section) {
                SettingsSection.PROFILE -> loadProfile()
                SettingsSection.PREFERENCES -> loadPreferences()
                SettingsSection.METADATA -> loadMetadata()
                SettingsSection.SESSIONS -> loadSessions()
                SettingsSection.SUBTITLES -> loadSubtitles()
                SettingsSection.PLAYBACK -> loadPlayback()
                SettingsSection.CINEMA -> loadCinema()
                SettingsSection.PRETRANSCODE -> loadPretranscode()
                SettingsSection.MARKERS -> loadMarkerStats()
                SettingsSection.DASHBOARD -> loadDashboard()
                SettingsSection.LIBRARIES -> loadLibraries()
                SettingsSection.ACTIVITY -> loadActivity()
                SettingsSection.USERS -> loadUsers()
                SettingsSection.WEBHOOKS -> loadWebhooks()
                SettingsSection.TASKS -> loadTasks()
                else -> { /* Other sections can be loaded lazily */ }
            }

            _uiState.update { it.copy(isLoading = false) }
        }
    }

    private suspend fun loadCinema() {
        try {
            val response = veloxApi.getCinemaSettings()
            if (response.isSuccessful) {
                _uiState.update { it.copy(adminCinemaSettings = response.body()?.data) }
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    fun updateCinemaSettings(enabled: Boolean? = null, maxTrailers: String? = null) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updateCinemaSettings(com.velox.app.data.model.UpdateCinemaRequest(enabled, maxTrailers))
                if (res.isSuccessful) {
                    _uiState.update { it.copy(adminCinemaSettings = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    private suspend fun loadPretranscode() {
        try {
            val settingsRes = veloxApi.getPretranscodeSettings()
            if (settingsRes.isSuccessful) {
                _uiState.update { it.copy(pretranscodeSettings = settingsRes.body()?.data) }
            }
            
            val profilesRes = veloxApi.getPretranscodeProfiles()
            if (profilesRes.isSuccessful) {
                _uiState.update { it.copy(pretranscodeProfiles = profilesRes.body()?.data ?: emptyList()) }
            }

            refreshPretranscodeStatus()
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    private var pretranscodePollingJob: kotlinx.coroutines.Job? = null

    init {
        // ... (Keep existing init logic if any, else add polling here or tie to Tab Select)
        viewModelScope.launch {
            _uiState.map { it.selectedTab }.distinctUntilChanged().collect { tab ->
                if (tab == "pretranscode") {
                    startPretranscodePolling()
                } else {
                    stopPretranscodePolling()
                }
            }
        }
    }

    private fun startPretranscodePolling() {
        if (pretranscodePollingJob?.isActive == true) return
        pretranscodePollingJob = viewModelScope.launch {
            while (true) {
                refreshPretranscodeStatus()
                kotlinx.coroutines.delay(5000)
            }
        }
    }

    private fun stopPretranscodePolling() {
        pretranscodePollingJob?.cancel()
        pretranscodePollingJob = null
    }

    private suspend fun refreshPretranscodeStatus() {
        try {
            val statusRes = veloxApi.getPretranscodeStatus()
            if (statusRes.isSuccessful) {
                _uiState.update { it.copy(pretranscodeStatus = statusRes.body()?.data) }
            }
        } catch (e: Exception) {
            e.printStackTrace() // Ignore network errors during polling
        }
    }

    fun updatePretranscodeSettings(enabled: Boolean? = null, schedule: String? = null, concurrency: String? = null) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updatePretranscodeSettings(com.velox.app.data.model.UpdatePretranscodeRequest(enabled, schedule, concurrency))
                if (res.isSuccessful) {
                    _uiState.update { it.copy(pretranscodeSettings = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun togglePretranscodeProfile(id: Int, enabled: Boolean) {
        viewModelScope.launch {
            try {
                veloxApi.togglePretranscodeProfile(id, mapOf("enabled" to enabled))
                // Refresh local optimism
                _uiState.update { state -> 
                    state.copy(pretranscodeProfiles = state.pretranscodeProfiles.map { profile -> if (profile.id == id) profile.copy(enabled = enabled) else profile })
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun loadPretranscodeEstimate(libraryId: Int) {
        if (libraryId == 0) {
            _uiState.update { it.copy(pretranscodeEstimate = null) }
            return
        }
        viewModelScope.launch {
            try {
                val res = veloxApi.getPretranscodeEstimate(libraryId)
                if (res.isSuccessful) {
                    _uiState.update { it.copy(pretranscodeEstimate = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun executePretranscodeAction(action: String) {
        viewModelScope.launch {
            try {
                when (action) {
                    "start" -> veloxApi.startPretranscode()
                    "stop" -> veloxApi.stopPretranscode()
                    "resume" -> veloxApi.resumePretranscode()
                    "cleanup" -> veloxApi.cleanupPretranscodeFiles()
                }
                refreshPretranscodeStatus()
                if (action == "start" || action == "cleanup") {
                    // Update settings to reflect changes
                    loadPretranscode()
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    private suspend fun loadMarkerStats() {
        try {
            // Load libraries so the detection dropdown has data
            loadLibraries()
            val response = veloxApi.getMarkerStats()
            if (response.isSuccessful) {
                _uiState.update { it.copy(markerStats = response.body()?.data) }
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    fun runMarkerDetection(libraryId: Int) {
        viewModelScope.launch {
            _uiState.update { it.copy(isRunningDetection = true) }
            try {
                veloxApi.startMarkerBackfill(com.velox.app.data.model.BackfillMarkersRequest(libraryId = libraryId))
                _uiState.update { it.copy(successMessage = "Started fingerprint detection in background.") }
                // Optionally reload stats after a delay
                kotlinx.coroutines.delay(2000)
                loadMarkerStats()
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to start detection: ${e.message}") }
            } finally {
                _uiState.update { it.copy(isRunningDetection = false) }
            }
        }
    }

    private suspend fun loadProfile() {
        // Profile loaded from user info
    }

    private suspend fun loadPreferences() {
        // Load preferences from API
        try {
            val response = veloxApi.getPreferences()
            if (response.isSuccessful) {
                response.body()?.data?.let { prefs ->
                    _uiState.update { it.copy(preferences = UserPreferences(
                        subtitleLanguage = prefs.subtitleLanguage,
                        audioLanguage = prefs.audioLanguage,
                        maxStreamingQuality = prefs.maxStreamingQuality,
                        theme = prefs.theme,
                        language = prefs.language,
                    )) }
                }
            }
        } catch (e: Exception) {
            // Use defaults on error
        }
    }

    private suspend fun loadSessions() {
        try {
            val response = veloxApi.getSessions()
            if (response.isSuccessful) {
                response.body()?.data?.let { sessions ->
                    _uiState.update { it.copy(sessions = sessions.map { s ->
                        Session(
                            id = s.id,
                            deviceName = s.deviceName?.takeIf { it.isNotBlank() } ?: s.ipAddress ?: "Unknown Device",
                            ipAddress = s.ipAddress ?: "Unknown IP",
                            lastActive = s.lastActiveAt,
                            isCurrent = false, // Need actual current session info
                        )
                    }) }
                }
            }
        } catch (e: Exception) {
            // Use empty on error
        }
    }

    private suspend fun loadNotifications() {
        _uiState.update { it.copy(isLoading = true) }
        try {
            val response = veloxApi.getNotifications(limit = 50, offset = 0)
            if (response.isSuccessful) {
                response.body()?.data?.let { wrapper ->
                    _uiState.update { 
                        it.copy(
                            notificationsList = wrapper.notifications,
                            unreadNotificationsCount = wrapper.unreadCount
                        ) 
                    }
                }
            }
        } catch (e: Exception) {
            // Ignore for now
        }
        _uiState.update { it.copy(isLoading = false) }
    }
    private suspend fun loadSubtitles() {
        try {
            val openSubsRes = veloxApi.getOpenSubsSettings()
            val subdlRes = veloxApi.getProviderSettings("subdl")
            val deepLRes = veloxApi.getProviderSettings("deepl")
            val aiTranslationRes = veloxApi.getAITranslationSettings()
            val autoSubRes = veloxApi.getAutoSubSettings()

            _uiState.update { state ->
                state.copy(
                    openSubsSettings = openSubsRes.body()?.data,
                    subdlSettings = subdlRes.body()?.data,
                    deepLSettings = deepLRes.body()?.data,
                    aiTranslationSettings = aiTranslationRes.body()?.data,
                    autoSubSettings = autoSubRes.body()?.data,
                )
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    fun updateOpenSubsSettings(request: com.velox.app.data.model.UpdateOpenSubsRequest) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updateOpenSubsSettings(request)
                if (res.isSuccessful) {
                    _uiState.update { it.copy(openSubsSettings = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun updateAutoSubSettings(languages: String) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updateAutoSubSettings(com.velox.app.data.model.UpdateAutoSubRequest(languages))
                if (res.isSuccessful) {
                    _uiState.update { it.copy(autoSubSettings = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun updateAITranslationSettings(provider: String, apiKey: String, baseUrl: String, model: String) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updateAITranslationSettings(
                    com.velox.app.data.model.UpdateAITranslationRequest(
                        provider = provider,
                        apiKey = apiKey,
                        baseUrl = baseUrl,
                        model = model,
                    )
                )
                if (res.isSuccessful) {
                    _uiState.update { it.copy(aiTranslationSettings = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    private suspend fun loadPlayback() {
        try {
            val response = veloxApi.getPlaybackSettings()
            if (response.isSuccessful) {
                _uiState.update { it.copy(adminPlaybackSettings = response.body()?.data) }
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    fun updatePlaybackMode(mode: String) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updatePlaybackSettings(com.velox.app.data.model.UpdatePlaybackRequest(mode))
                if (res.isSuccessful) {
                    _uiState.update { it.copy(adminPlaybackSettings = res.body()?.data) }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    private suspend fun loadDashboard() {
        try {
            val serverResponse = veloxApi.getServerInfo()
            if (serverResponse.isSuccessful) {
                serverResponse.body()?.data?.let { info ->
                    _uiState.update { it.copy(serverInfo = info) }
                }
            }

            val libraryResponse = veloxApi.getLibraryStats()
            if (libraryResponse.isSuccessful) {
                libraryResponse.body()?.data?.let { stats ->
                    _uiState.update { it.copy(libraryStats = stats) }
                }
            }

            val activityResponse = veloxApi.getAdminActivity(limit = 10)
            if (activityResponse.isSuccessful) {
                activityResponse.body()?.data?.let { activities ->
                    _uiState.update { it.copy(activityLogs = activities) }
                }
            }
        } catch (e: Exception) {
            // Use defaults on error
        }
    }

    private suspend fun loadLibraries() {
        mediaRepository.getLibraries().onSuccess { libraries ->
            _uiState.update { state ->
                state.copy(libraries = libraries)
            }
        }
    }

    private suspend fun loadActivity() {
        try {
            val response = veloxApi.getAdminActivity(limit = 50, offset = 0)
            if (response.isSuccessful) {
                response.body()?.data?.let { logs ->
                    _uiState.update { it.copy(activityLogs = logs) }
                }
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    private suspend fun loadUsers() {
        try {
            val response = veloxApi.getAdminUsers()
            if (response.isSuccessful) {
                response.body()?.data?.let { users ->
                    _uiState.update { it.copy(users = users.map { u ->
                        AdminUser(
                            id = u.id,
                            username = u.username,
                            displayName = u.displayName,
                            isAdmin = u.isAdmin,
                            lastActive = u.lastActive,
                            createdAt = u.createdAt,
                        )
                    }) }
                }
            }
        } catch (e: Exception) {
            // Use empty on error
        }
    }

    private suspend fun loadWebhooks() {
        try {
            val response = veloxApi.getWebhooks()
            if (response.isSuccessful) {
                response.body()?.data?.let { webhooks ->
                    _uiState.update { it.copy(webhooks = webhooks.map { w ->
                        Webhook(
                            id = w.id,
                            url = w.url,
                            events = w.events,
                            active = w.active,
                        )
                    }) }
                }
            }
        } catch (e: Exception) {
            // Use empty on error
        }
    }

    private suspend fun loadTasks() {
        try {
            val response = veloxApi.getTasks()
            if (response.isSuccessful) {
                response.body()?.data?.let { ts ->
                    _uiState.update { it.copy(tasks = ts) }
                }
            }
        } catch (e: Exception) {
            // Use empty on error
        }
    }

    fun runTask(name: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(runningTaskName = name) }
            try {
                // We use name but API might need the ID? No, endpoint is /run but web uses `runTask(name)`. Wait, VeloxApi.kt has `POST admin/tasks/{id}/run`. Let's check it.
                // Wait, if VeloxApi has id as Int, I might need to adjust it or pass Name string. I'll just adjust VeloxApi in next step.
                veloxApi.runTask(name)
                // Reload after finishing
                loadTasks()
            } catch (e: Exception) {
                // Failure
            } finally {
                _uiState.update { it.copy(runningTaskName = null) }
            }
        }
    }

    fun checkAppUpdate() {
        viewModelScope.launch {
            _uiState.update { it.copy(isCheckingUpdate = true, updateStatus = "Checking...") }
            try {
                val currentVersionCode = com.velox.app.BuildConfig.VERSION_CODE
                val response = veloxApi.getLatestAppVersion("android")
                
                if (response.isSuccessful) {
                    val data = response.body()
                    if (data != null && data.error == null) {
                        if (data.versionCode > currentVersionCode) {
                            _uiState.update {
                                it.copy(
                                    isCheckingUpdate = false,
                                    latestVersion = data.versionName,
                                    updateDownloadUrl = data.downloadUrl,
                                    updateIsMandatory = data.isMandatory,
                                    updateStatus = if (data.isMandatory) "Breaking change! Mandatory update required: ${data.versionName}" 
                                                  else "New version available: ${data.versionName}"
                                )
                            }
                        } else {
                            _uiState.update { it.copy(isCheckingUpdate = false, updateStatus = "App is up to date") }
                        }
                    } else {
                        // Backend returned 404 handled as success maybe? No, 404 is response.isSuccessful == false
                        _uiState.update { it.copy(isCheckingUpdate = false, updateStatus = "App is up to date (or none found)") }
                    }
                } else if (response.code() == 404) {
                    _uiState.update { it.copy(isCheckingUpdate = false, updateStatus = "App is up to date") }
                } else {
                    _uiState.update { it.copy(isCheckingUpdate = false, updateStatus = "Update check failed (Code: ${response.code()})") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isCheckingUpdate = false, updateStatus = "Error checking update: ${e.message}") }
            }
        }
    }

    fun updateTaskInterval(name: String, interval: String) {
        viewModelScope.launch {
            try {
                val response = veloxApi.updateTaskInterval(name, com.velox.app.data.model.UpdateTaskIntervalRequest(interval))
                if (response.isSuccessful) {
                    loadTasks()
                    _uiState.update { it.copy(successMessage = "Task interval updated") }
                } else {
                    _uiState.update { it.copy(error = "Failed to update interval") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to update interval: ${e.message}") }
            }
        }
    }

    // Profile actions
    fun updateDisplayName(name: String) {
        _uiState.update { it.copy(displayName = name) }
    }

    fun saveProfile() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = veloxApi.updateProfile(mapOf("display_name" to _uiState.value.displayName))
                if (response.isSuccessful) {
                    authRepository.getMe() // Refresh the local user state
                    _uiState.update { it.copy(successMessage = "Profile updated successfully") }
                } else {
                    _uiState.update { it.copy(error = "Failed to update profile") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to update profile: ${e.message}") }
            }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    // Preferences actions
    fun updatePreference(field: String, value: String) {
        _uiState.update { state ->
            val p = state.preferences
            state.copy(
                preferences = when (field) {
                    "subtitleLanguage" -> p.copy(subtitleLanguage = value)
                    "audioLanguage" -> p.copy(audioLanguage = value)
                    "maxQuality" -> p.copy(maxStreamingQuality = value)
                    "theme" -> p.copy(theme = value)
                    "language" -> p.copy(language = value)
                    else -> p
                }
            )
        }
    }

    fun savePreferences() {
        val user = _uiState.value.user ?: return
        val prefs = _uiState.value.preferences
        
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, successMessage = null) }
            try {
                val response = veloxApi.updatePreferences(
                    com.velox.app.data.model.UpdatePreferencesRequest(
                        userId = user.id,
                        subtitleLanguage = prefs.subtitleLanguage,
                        audioLanguage = prefs.audioLanguage,
                        maxStreamingQuality = prefs.maxStreamingQuality,
                        theme = prefs.theme,
                        language = prefs.language,
                    )
                )
                if (response.isSuccessful) {
                    _uiState.update { it.copy(successMessage = "Preferences updated successfully") }
                } else {
                    _uiState.update { it.copy(error = "Failed to update preferences") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to update preferences: ${e.message}") }
            }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    // Security actions
    fun updateCurrentPassword(password: String) {
        _uiState.update { it.copy(currentPassword = password) }
    }

    fun updateNewPassword(password: String) {
        _uiState.update { it.copy(newPassword = password) }
    }

    fun updateConfirmPassword(password: String) {
        _uiState.update { it.copy(confirmPassword = password) }
    }

    fun changePassword() {
        val state = _uiState.value
        if (state.newPassword != state.confirmPassword) {
            _uiState.update { it.copy(securityError = "Passwords do not match") }
            return
        }
        if (state.newPassword.length < 8) {
            _uiState.update { it.copy(securityError = "Password must be at least 8 characters") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, securityError = null) }
            try {
                val response = veloxApi.changePassword(ChangePasswordRequest(
                    oldPassword = state.currentPassword,
                    newPassword = state.newPassword,
                ))
                if (response.isSuccessful) {
                    _uiState.update {
                        it.copy(
                            securitySuccess = "Password changed successfully",
                            currentPassword = "",
                            newPassword = "",
                            confirmPassword = "",
                        )
                    }
                } else {
                    _uiState.update { it.copy(securityError = "Failed to change password") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(securityError = "Failed to change password: ${e.message}") }
            }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    // Sessions actions
    fun revokeSession(sessionId: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.revokeSession(sessionId)
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        state.copy(sessions = state.sessions.filter { it.id != sessionId })
                    }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(sessionRevokeError = "Failed to revoke session") }
            }
        }
    }

    // Notifications actions
    fun loadNotificationsAction() {
        viewModelScope.launch { loadNotifications() }
    }

    fun markNotificationRead(id: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.markNotificationsRead(
                    com.velox.app.data.model.MarkReadRequest(listOf(id))
                )
                if (response.isSuccessful) {
                    loadNotifications()
                }
            } catch (e: Exception) {
                // Ignore
            }
        }
    }

    fun markAllNotificationsRead() {
        viewModelScope.launch {
            try {
                val response = veloxApi.markAllNotificationsRead()
                if (response.isSuccessful) {
                    loadNotifications()
                }
            } catch (e: Exception) {
                // Ignore
            }
        }
    }

    fun deleteNotification(id: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.deleteNotifications(
                    com.velox.app.data.model.DeleteNotificationsRequest(listOf(id))
                )
                if (response.isSuccessful) {
                    loadNotifications()
                }
            } catch (e: Exception) {
                // Ignore
            }
        }
    }

    // Subtitles actions replaced by Admin Subtitles actions

    // Playback actions replaced by Admin Playback Actions

    // Cinema actions replaced by Admin Cinema Actions

    // Skip markers actions
    fun updateSkipMarkerSetting(type: String, enabled: Boolean) {
        _uiState.update { state ->
            val markers = state.skipMarkers
            state.copy(
                skipMarkers = when (type) {
                    "skipIntro" -> markers.copy(skipIntro = enabled)
                    "skipCredits" -> markers.copy(skipCredits = enabled)
                    "autoSkipSponsor" -> markers.copy(autoSkipSponsor = enabled)
                    else -> markers
                },
            )
        }
    }



    // Admin - Webhook actions
    fun deleteWebhook(webhookId: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.deleteWebhook(webhookId)
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        state.copy(webhooks = state.webhooks.filter { it.id != webhookId })
                    }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to delete webhook") }
            }
        }
    }

    fun createWebhook(url: String, events: List<String>, active: Boolean) {
        viewModelScope.launch {
            try {
                val response = veloxApi.createWebhook(
                    CreateWebhookRequest(url = url, events = events, active = active)
                )
                if (response.isSuccessful) {
                    loadWebhooks()
                } else {
                    _uiState.update { it.copy(error = "Failed to create webhook") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to create webhook: ${e.message}") }
            }
        }
    }

    fun updateWebhook(webhookId: Int, url: String?, events: List<String>?, active: Boolean?) {
        viewModelScope.launch {
            try {
                val response = veloxApi.updateWebhook(
                    webhookId,
                    UpdateWebhookRequest(url = url, events = events, active = active)
                )
                if (response.isSuccessful) {
                    loadWebhooks()
                } else {
                    _uiState.update { it.copy(error = "Failed to update webhook") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to update webhook: ${e.message}") }
            }
        }
    }

    // Admin - Library actions
    fun createLibrary(name: String, type: String, paths: List<String>) {
        viewModelScope.launch {
            try {
                val response = veloxApi.createLibrary(
                    CreateLibraryRequest(name = name, type = type, paths = paths)
                )
                if (response.isSuccessful) {
                    loadLibraries()
                } else {
                    _uiState.update { it.copy(error = "Failed to create library") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to create library: ${e.message}") }
            }
        }
    }

    fun updateLibrary(libraryId: Int, name: String?, paths: List<String>?) {
        viewModelScope.launch {
            try {
                val response = veloxApi.updateLibrary(
                    libraryId,
                    UpdateLibraryRequest(name = name, paths = paths)
                )
                if (response.isSuccessful) {
                    loadLibraries()
                } else {
                    _uiState.update { it.copy(error = "Failed to update library") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to update library: ${e.message}") }
            }
        }
    }

    fun deleteLibrary(libraryId: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.deleteLibrary(libraryId)
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        state.copy(libraries = state.libraries.filter { it.id != libraryId })
                    }
                } else {
                    _uiState.update { it.copy(error = "Failed to delete library") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to delete library: ${e.message}") }
            }
        }
    }

    fun scanLibrary(libraryId: Int, force: Boolean = false) {
        viewModelScope.launch {
            _uiState.update { it.copy(scanningLibraryId = libraryId) }
            try {
                val response = veloxApi.scanLibrary(libraryId, force)
                if (response.isSuccessful) {
                    _uiState.update { it.copy(successMessage = "Library scan started") }
                } else {
                    _uiState.update { it.copy(error = "Failed to scan library") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to scan library: ${e.message}") }
            } finally {
                _uiState.update { it.copy(scanningLibraryId = null) }
            }
        }
    }

    // Admin - User actions
    fun createUser(username: String, password: String, displayName: String, isAdmin: Boolean) {
        viewModelScope.launch {
            try {
                val response = veloxApi.createUser(
                    CreateUserRequest(
                        username = username,
                        password = password,
                        displayName = displayName,
                        isAdmin = isAdmin,
                    )
                )
                if (response.isSuccessful) {
                    loadUsers()
                } else {
                    _uiState.update { it.copy(error = "Failed to create user") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to create user: ${e.message}") }
            }
        }
    }

    fun updateUser(userId: Int, displayName: String?, isAdmin: Boolean?) {
        viewModelScope.launch {
            try {
                val response = veloxApi.updateUser(
                    userId,
                    UpdateUserRequest(displayName = displayName, isAdmin = isAdmin)
                )
                if (response.isSuccessful) {
                    loadUsers()
                } else {
                    _uiState.update { it.copy(error = "Failed to update user") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to update user: ${e.message}") }
            }
        }
    }

    fun deleteUser(userId: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.deleteUser(userId)
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        state.copy(users = state.users.filter { it.id != userId })
                    }
                } else {
                    _uiState.update { it.copy(error = "Failed to delete user") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to delete user: ${e.message}") }
            }
        }
    }

    fun clearMessages() {
        _uiState.update { it.copy(error = null, successMessage = null, securityError = null, securitySuccess = null, sessionRevokeError = null) }
    }

    private suspend fun loadMetadata() {
        try {
            val tmdbRes = veloxApi.getProviderSettings("tmdb")
            val omdbRes = veloxApi.getProviderSettings("omdb")
            val tvdbRes = veloxApi.getProviderSettings("tvdb")
            val fanartRes = veloxApi.getProviderSettings("fanart")

            _uiState.update { state ->
                state.copy(
                    tmdbSettings = tmdbRes.body()?.data,
                    omdbSettings = omdbRes.body()?.data,
                    tvdbSettings = tvdbRes.body()?.data,
                    fanartSettings = fanartRes.body()?.data,
                )
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    fun updateProviderApiKey(provider: String, apiKey: String) {
        viewModelScope.launch {
            try {
                val res = veloxApi.updateProviderSettings(provider, com.velox.app.data.model.UpdateProviderRequest(apiKey))
                if (res.isSuccessful) {
                    val settings = res.body()?.data
                    _uiState.update { state ->
                        when(provider) {
                            "tmdb" -> state.copy(tmdbSettings = settings)
                            "omdb" -> state.copy(omdbSettings = settings)
                            "tvdb" -> state.copy(tvdbSettings = settings)
                            "fanart" -> state.copy(fanartSettings = settings)
                            else -> state
                        }
                    }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun refreshAllMetadata() {
        viewModelScope.launch {
            try {
                veloxApi.refreshAllMetadata()
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

}
