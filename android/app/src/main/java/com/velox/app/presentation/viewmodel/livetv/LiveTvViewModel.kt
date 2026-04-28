package com.velox.app.presentation.viewmodel.livetv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.AuthManager
import com.velox.app.data.api.ServerPrefsManager
import com.velox.app.data.model.UpdatePreferencesRequest
import com.velox.app.data.model.UserPreferencesDto
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.domain.model.livetv.LiveGroup
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.domain.repository.SettingsRepository
import com.velox.app.domain.repository.livetv.LiveTvRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.joinAll
import kotlinx.coroutines.launch
import timber.log.Timber
import java.net.URLEncoder
import javax.inject.Inject

@HiltViewModel
class LiveTvViewModel @Inject constructor(
    private val repository: LiveTvRepository,
    private val settingsRepository: SettingsRepository,
    serverPrefsManager: ServerPrefsManager,
    authManager: AuthManager,
) : ViewModel() {

    companion object {
        const val GROUP_ALL = "all"
        const val GROUP_HIDDEN = "Hidden"
    }

    private val _groups = MutableStateFlow<DataResult<List<LiveGroup>>?>(null)
    val groups: StateFlow<DataResult<List<LiveGroup>>?> = _groups.asStateFlow()

    private val _channels = MutableStateFlow<DataResult<List<LiveChannel>>?>(null)
    val channels: StateFlow<DataResult<List<LiveChannel>>?> = _channels.asStateFlow()

    private val _selectedGroup = MutableStateFlow(GROUP_ALL)
    val selectedGroup: StateFlow<String> = _selectedGroup.asStateFlow()

    private val _playingChannel = MutableStateFlow<LiveChannel?>(null)
    val playingChannel: StateFlow<LiveChannel?> = _playingChannel.asStateFlow()

    private val _epgPrograms = MutableStateFlow<List<LiveProgram>>(emptyList())
    val epgPrograms: StateFlow<List<LiveProgram>> = _epgPrograms.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing.asStateFlow()

    val baseUrl: String = serverPrefsManager.getServerUrlSync()
    private val accessToken: String? = authManager.getAccessTokenSync()

    fun streamUrl(channelId: Int): String {
        val base = "${baseUrl}livetv/stream/$channelId"
        val token = accessToken?.takeIf { it.isNotBlank() } ?: return base
        return "$base?token=${URLEncoder.encode(token, "UTF-8")}"
    }

    private var cachedPrefs: UserPreferencesDto? = null
    private var lastLoadedEpgChannelId: Int? = null
    private var heroSeeded = false

    init {
        loadGroups()
        loadChannels(GROUP_ALL)
        loadPreferences()
    }

    fun selectGroup(group: String) {
        if (_selectedGroup.value == group) return
        _selectedGroup.value = group
        loadChannels(group)
    }

    fun setPlayingChannel(channel: LiveChannel) {
        if (_playingChannel.value?.id == channel.id) return
        _playingChannel.value = channel
        heroSeeded = true
        loadEpgIfNeeded(channel.id)
        persistLastWatched(channel)
    }

    fun persistLastWatched(channel: LiveChannel) {
        val prefs = cachedPrefs ?: return
        viewModelScope.launch {
            runCatching {
                val response = settingsRepository.updatePreferences(
                    UpdatePreferencesRequest(
                        userId = prefs.userId,
                        subtitleLanguage = prefs.subtitleLanguage,
                        audioLanguage = prefs.audioLanguage,
                        maxStreamingQuality = prefs.maxStreamingQuality,
                        theme = prefs.theme,
                        language = prefs.language,
                        lastLiveChannelId = channel.id.toLong(),
                    )
                )
                if (response.isSuccessful) {
                    response.body()?.data?.let { cachedPrefs = it }
                }
            }.onFailure { Timber.w(it, "Persist last-live-channel failed") }
        }
    }

    fun refresh() {
        if (_isRefreshing.value) return
        _isRefreshing.value = true
        viewModelScope.launch {
            val groupsJob = launch {
                repository.getLiveTvGroups().collect { _groups.value = it }
            }
            val filter = if (_selectedGroup.value == GROUP_ALL) null else _selectedGroup.value
            val channelsJob = launch {
                repository.getLiveTvChannels(filter).collect { result ->
                    _channels.value = result
                    if (result is DataResult.Success) seedHeroIfPossible(result.data)
                }
            }
            val epgJob = _playingChannel.value?.id?.let { channelId ->
                launch {
                    repository.getLiveTvEpg(channelId).collect { result ->
                        _epgPrograms.value = if (result is DataResult.Success) result.data else emptyList()
                    }
                }
            }
            joinAll(*listOfNotNull(groupsJob, channelsJob, epgJob).toTypedArray())
            _isRefreshing.value = false
        }
    }

    fun toggleHidden(channel: LiveChannel) {
        viewModelScope.launch {
            val result = repository.toggleChannelHidden(channel.id)
            if (result is DataResult.Success) {
                loadChannels(_selectedGroup.value)
                if (_playingChannel.value?.id == channel.id && result.data) {
                    // Hidden the currently playing channel; clear the hero
                    _playingChannel.value = null
                    heroSeeded = false
                    _epgPrograms.value = emptyList()
                    lastLoadedEpgChannelId = null
                }
            }
        }
    }

    private fun loadGroups() {
        viewModelScope.launch {
            repository.getLiveTvGroups().collect { _groups.value = it }
        }
    }

    private fun loadChannels(group: String) {
        _channels.value = null
        val filter = when (group) {
            GROUP_ALL -> null
            else -> group
        }
        viewModelScope.launch {
            repository.getLiveTvChannels(filter).collect { result ->
                _channels.value = result
                if (result is DataResult.Success) {
                    seedHeroIfPossible(result.data)
                }
            }
        }
    }

    private fun seedHeroIfPossible(channels: List<LiveChannel>) {
        if (heroSeeded || channels.isEmpty()) return
        val prefId = cachedPrefs?.lastLiveChannelId
        val preferred = prefId?.let { id -> channels.firstOrNull { it.id.toLong() == id } }
        val seeded = preferred ?: channels.first()
        _playingChannel.value = seeded
        heroSeeded = true
        loadEpgIfNeeded(seeded.id)
    }

    private fun loadEpgIfNeeded(channelId: Int) {
        if (lastLoadedEpgChannelId == channelId) return
        lastLoadedEpgChannelId = channelId
        viewModelScope.launch {
            repository.getLiveTvEpg(channelId).collect { result ->
                _epgPrograms.value = if (result is DataResult.Success) result.data else emptyList()
            }
        }
    }

    private fun loadPreferences() {
        viewModelScope.launch {
            runCatching {
                val response = settingsRepository.getPreferences()
                if (response.isSuccessful) {
                    cachedPrefs = response.body()?.data
                    (_channels.value as? DataResult.Success)?.data?.let { seedHeroIfPossible(it) }
                }
            }.onFailure { Timber.w(it, "Load preferences failed") }
        }
    }
}
