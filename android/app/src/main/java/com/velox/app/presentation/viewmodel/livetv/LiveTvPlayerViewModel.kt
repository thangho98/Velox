package com.velox.app.presentation.viewmodel.livetv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.AuthManager
import com.velox.app.data.api.ServerPrefsManager
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.domain.repository.livetv.LiveTvRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.net.URLEncoder
import javax.inject.Inject

@HiltViewModel
class LiveTvPlayerViewModel @Inject constructor(
    serverPrefsManager: ServerPrefsManager,
    authManager: AuthManager,
    private val repository: LiveTvRepository,
) : ViewModel() {

    val baseUrl: String = serverPrefsManager.getServerUrlSync()
    private val accessToken: String? = authManager.getAccessTokenSync()

    fun streamUrl(channelId: Int): String {
        val base = "${baseUrl}livetv/stream/$channelId"
        val token = accessToken?.takeIf { it.isNotBlank() } ?: return base
        return "$base?token=${URLEncoder.encode(token, "UTF-8")}"
    }

    private val _channel = MutableStateFlow<LiveChannel?>(null)
    val channel: StateFlow<LiveChannel?> = _channel.asStateFlow()

    private val _epg = MutableStateFlow<List<LiveProgram>>(emptyList())
    val epg: StateFlow<List<LiveProgram>> = _epg.asStateFlow()

    private var loadedChannelId: Int? = null

    fun load(channelId: Int) {
        if (loadedChannelId == channelId) return
        loadedChannelId = channelId
        viewModelScope.launch {
            repository.getLiveTvChannel(channelId).collect { result ->
                if (result is DataResult.Success) _channel.value = result.data
            }
        }
        viewModelScope.launch {
            repository.getLiveTvEpg(channelId).collect { result ->
                if (result is DataResult.Success) _epg.value = result.data
            }
        }
    }
}
