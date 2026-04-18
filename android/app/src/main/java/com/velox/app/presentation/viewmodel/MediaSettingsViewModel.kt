package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.model.*
import com.velox.app.domain.model.Library
import com.velox.app.domain.repository.MediaRepository
import com.velox.app.domain.repository.SettingsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import timber.log.Timber
import javax.inject.Inject

data class MediaSettingsUiState(
    val isLoading: Boolean = false,

    // Subtitles
    val openSubsSettings: OpenSubsSettingsDto? = null,
    val subdlSettings: ProviderSettingsDto? = null,
    val deepLSettings: ProviderSettingsDto? = null,
    val aiTranslationSettings: AITranslationSettingsDto? = null,
    val autoSubSettings: AutoSubSettingsDto? = null,
    val autoTranslateSettings: AutoTranslateSettingsDto? = null,

    // Playback
    val adminPlaybackSettings: PlaybackSettingsDto? = null,

    // Cinema
    val adminCinemaSettings: CinemaSettingsDto? = null,

    // Pre-transcode
    val pretranscodeSettings: PretranscodeSettingsDto? = null,
    val pretranscodeStatus: PretranscodeStatusDto? = null,
    val pretranscodeProfiles: List<PretranscodeProfileDto> = emptyList(),
    val pretranscodeEstimate: StorageEstimateDto? = null,

    // Markers Dashboard
    val markerStats: MarkerStatsDto? = null,
    val isRunningDetection: Boolean = false,

    // Metadata
    val anilistSettings: ProviderSettingsDto? = null,
    val tmdbSettings: ProviderSettingsDto? = null,
    val omdbSettings: ProviderSettingsDto? = null,
    val tvdbSettings: ProviderSettingsDto? = null,
    val fanartSettings: ProviderSettingsDto? = null,

    // Shared Admin State
    val libraries: List<Library> = emptyList(),

    val error: String? = null,
    val successMessage: String? = null,
)

@HiltViewModel
class MediaSettingsViewModel @Inject constructor(
    private val settingsRepository: SettingsRepository,
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(MediaSettingsUiState())
    val uiState: StateFlow<MediaSettingsUiState> = _uiState.asStateFlow()

    private var pretranscodePollingJob: kotlinx.coroutines.Job? = null

    fun loadSubtitles() {
        viewModelScope.launch {
            val openSubs = settingsRepository.fetchOpenSubsSettings().getOrNull()
            val subdl = settingsRepository.fetchProviderSettings("subdl").getOrNull()
            val deepL = settingsRepository.fetchProviderSettings("deepl").getOrNull()
            val aiTranslation = settingsRepository.fetchAITranslationSettings().getOrNull()
            val autoSub = settingsRepository.fetchAutoSubSettings().getOrNull()
            val autoTranslate = settingsRepository.fetchAutoTranslateSettings().getOrNull()

            _uiState.update { state ->
                state.copy(
                    openSubsSettings = openSubs,
                    subdlSettings = subdl,
                    deepLSettings = deepL,
                    aiTranslationSettings = aiTranslation,
                    autoSubSettings = autoSub,
                    autoTranslateSettings = autoTranslate,
                )
            }
        }
    }

    fun loadPlayback() {
        viewModelScope.launch {
            settingsRepository.fetchPlaybackSettings()
                .onSuccess { data -> _uiState.update { it.copy(adminPlaybackSettings = data) } }
                .onError { Timber.e("loadPlayback: ${it.message}") }
        }
    }

    fun loadCinema() {
        viewModelScope.launch {
            settingsRepository.fetchCinemaSettings()
                .onSuccess { data -> _uiState.update { it.copy(adminCinemaSettings = data) } }
                .onError { Timber.e("loadCinema: ${it.message}") }
        }
    }

    fun loadPretranscode() {
        viewModelScope.launch {
            settingsRepository.fetchPretranscodeSettings()
                .onSuccess { data -> _uiState.update { it.copy(pretranscodeSettings = data) } }

            settingsRepository.fetchPretranscodeProfiles()
                .onSuccess { profiles -> _uiState.update { it.copy(pretranscodeProfiles = profiles) } }

            refreshPretranscodeStatus()
        }
    }

    fun startPretranscodePolling() {
        if (pretranscodePollingJob?.isActive == true) return
        pretranscodePollingJob = viewModelScope.launch {
            while (true) {
                refreshPretranscodeStatus()
                kotlinx.coroutines.delay(5000)
            }
        }
    }

    fun stopPretranscodePolling() {
        pretranscodePollingJob?.cancel()
        pretranscodePollingJob = null
    }

    suspend fun refreshPretranscodeStatus() {
        settingsRepository.fetchPretranscodeStatus()
            .onSuccess { data -> _uiState.update { it.copy(pretranscodeStatus = data) } }
    }

    fun loadMarkerStats() {
        viewModelScope.launch {
            settingsRepository.fetchMarkerStats()
                .onSuccess { data -> _uiState.update { it.copy(markerStats = data) } }
                .onError { Timber.e("loadMarkerStats: ${it.message}") }
        }
    }

    fun loadMetadata() {
        viewModelScope.launch {
            val anilist = settingsRepository.fetchProviderSettings("anilist").getOrNull()
            val tmdb = settingsRepository.fetchProviderSettings("tmdb").getOrNull()
            val omdb = settingsRepository.fetchProviderSettings("omdb").getOrNull()
            val tvdb = settingsRepository.fetchProviderSettings("tvdb").getOrNull()
            val fanart = settingsRepository.fetchProviderSettings("fanart").getOrNull()

            _uiState.update { state ->
                state.copy(
                    anilistSettings = anilist,
                    tmdbSettings = tmdb,
                    omdbSettings = omdb,
                    tvdbSettings = tvdb,
                    fanartSettings = fanart,
                )
            }
        }
    }

    fun loadLibraries() {
        viewModelScope.launch {
            mediaRepository.getLibraries()
                .onSuccess { data -> _uiState.update { it.copy(libraries = data) } }
                .onFailure { error -> Timber.e("loadLibraries: ${error.message}") }
        }
    }

    fun refreshAllMetadata() {
        viewModelScope.launch {
            settingsRepository.refreshAllMetadataSafe()
                .onSuccess {
                    _uiState.update { it.copy(successMessage = "Metadata refresh started") }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to refresh metadata: ${error.message}") }
                }
        }
    }

    fun updateProviderApiKey(provider: String, apiKey: String) {
        viewModelScope.launch {
            settingsRepository.updateProviderSettingsSafe(provider, UpdateProviderRequest(apiKey))
                .onSuccess { settings ->
                    _uiState.update { state ->
                        when (provider) {
                            "anilist" -> state.copy(anilistSettings = settings)
                            "tmdb" -> state.copy(tmdbSettings = settings)
                            "omdb" -> state.copy(omdbSettings = settings)
                            "tvdb" -> state.copy(tvdbSettings = settings)
                            "fanart" -> state.copy(fanartSettings = settings)
                            "subdl" -> state.copy(subdlSettings = settings)
                            "deepl" -> state.copy(deepLSettings = settings)
                            else -> state
                        }
                    }
                    _uiState.update { it.copy(successMessage = "Provider updated successfully") }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update provider settings: ${error.message}") }
                }
        }
    }

    fun updateOpenSubsSettings(request: UpdateOpenSubsRequest) {
        viewModelScope.launch {
            settingsRepository.updateOpenSubsSettingsSafe(request)
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            openSubsSettings = data,
                            successMessage = "OpenSubtitles settings updated",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update OpenSubtitles settings: ${error.message}") }
                }
        }
    }

    fun updateAITranslationSettings(
        provider: String,
        apiKey: String,
        baseUrl: String,
        model: String,
    ) {
        viewModelScope.launch {
            settingsRepository.updateAITranslationSettingsSafe(
                UpdateAITranslationRequest(
                    provider = provider,
                    apiKey = apiKey,
                    baseUrl = baseUrl,
                    model = model,
                )
            )
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            aiTranslationSettings = data,
                            successMessage = "AI translation settings updated",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update AI translation settings: ${error.message}") }
                }
        }
    }

    fun updateAutoSubSettings(languages: String) {
        viewModelScope.launch {
            settingsRepository.updateAutoSubSettingsSafe(UpdateAutoSubRequest(languages))
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            autoSubSettings = data,
                            successMessage = "Auto subtitle settings updated",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update auto subtitle settings: ${error.message}") }
                }
        }
    }

    fun updatePlaybackMode(mode: String) {
        viewModelScope.launch {
            settingsRepository.savePlaybackSettings(UpdatePlaybackRequest(mode))
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            adminPlaybackSettings = data,
                            successMessage = "Playback mode updated",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update playback mode: ${error.message}") }
                }
        }
    }

    fun updateCinemaSettings(enabled: Boolean? = null, maxTrailers: String? = null) {
        viewModelScope.launch {
            settingsRepository.saveCinemaSettings(
                UpdateCinemaRequest(enabled = enabled, maxTrailers = maxTrailers)
            )
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            adminCinemaSettings = data,
                            successMessage = "Cinema settings updated",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update cinema settings: ${error.message}") }
                }
        }
    }

    fun loadPretranscodeEstimate(libraryId: Int) {
        if (libraryId <= 0) {
            _uiState.update { it.copy(pretranscodeEstimate = null) }
            return
        }

        viewModelScope.launch {
            settingsRepository.fetchPretranscodeEstimate(libraryId)
                .onSuccess { data -> _uiState.update { it.copy(pretranscodeEstimate = data) } }
                .onError { Timber.e("loadPretranscodeEstimate: ${it.message}") }
        }
    }

    fun updatePretranscodeSettings(
        enabled: Boolean? = null,
        schedule: String? = null,
        concurrency: String? = null,
    ) {
        viewModelScope.launch {
            settingsRepository.savePretranscodeSettings(
                UpdatePretranscodeRequest(
                    enabled = enabled,
                    schedule = schedule,
                    concurrency = concurrency,
                )
            )
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            pretranscodeSettings = data,
                            successMessage = "Pre-transcode settings updated",
                        )
                    }
                    refreshPretranscodeStatus()
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update pre-transcode settings: ${error.message}") }
                }
        }
    }

    fun togglePretranscodeProfile(profileId: Int, enabled: Boolean) {
        viewModelScope.launch {
            settingsRepository.togglePretranscodeProfileSafe(profileId, mapOf("enabled" to enabled))
                .onSuccess { loadPretranscode() }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update profile state: ${error.message}") }
                }
        }
    }

    fun executePretranscodeAction(action: String) {
        viewModelScope.launch {
            settingsRepository.executePretranscodeActionSafe(action)
                .onSuccess {
                    refreshPretranscodeStatus()
                    _uiState.update { it.copy(successMessage = "Action \"$action\" executed") }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to execute \"$action\": ${error.message}") }
                }
        }
    }

    fun runMarkerDetection(libraryId: Int) {
        viewModelScope.launch {
            _uiState.update { it.copy(isRunningDetection = true, error = null) }
            settingsRepository.startMarkerBackfillSafe(
                BackfillMarkersRequest(libraryId = libraryId.takeIf { it > 0 })
            )
                .onSuccess {
                    loadMarkerStats()
                    _uiState.update { it.copy(successMessage = "Marker detection started") }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to run marker detection: ${error.message}") }
                }
            _uiState.update { it.copy(isRunningDetection = false) }
        }
    }

    fun clearMessages() {
        _uiState.update { it.copy(error = null, successMessage = null) }
    }


    fun updateAutoTranslateSettings(enabled: Boolean, languages: String) {
        viewModelScope.launch {
            settingsRepository.updateAutoTranslateSettingsSafe(com.velox.app.data.model.UpdateAutoTranslateRequest(enabled, languages))
                .onSuccess { data ->
                    _uiState.update {
                        it.copy(
                            autoTranslateSettings = data,
                            successMessage = "Auto-translate settings updated",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Update failed: ${error.message}") }
                }
        }
    }
}
