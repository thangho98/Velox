package com.velox.app.presentation.viewmodel

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.VeloxApi
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.CinemaDto
import com.velox.app.data.model.TrailerDto
import com.velox.app.domain.model.MediaDetail
import com.velox.app.domain.model.MediaFile
import com.velox.app.domain.model.SubtitleTrack
import com.velox.app.domain.model.WatchProgress
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MediaDetailUiState(
    val isLoading: Boolean = true,
    val media: MediaDetail? = null,
    val progress: WatchProgress? = null,
    val isFavorite: Boolean = false,
    val isWatched: Boolean = false,
    val cinema: CinemaDto? = null,
    val error: String? = null,
    val mediaFiles: List<MediaFile> = emptyList(),
    val primaryFile: MediaFile? = null,
    val subtitleTracks: List<SubtitleTrack> = emptyList(),
    val selectedSubtitleLanguage: String? = null,
    val isAdmin: Boolean = false,
)

@HiltViewModel
class MediaDetailViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
    private val authRepository: com.velox.app.domain.repository.AuthRepository,
    private val apiProvider: VeloxApiProvider,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val veloxApi: VeloxApi
        get() = apiProvider.getApi()

    val mediaId: Int = checkNotNull(savedStateHandle["mediaId"])

    private val _uiState = MutableStateFlow(MediaDetailUiState())
    val uiState: StateFlow<MediaDetailUiState> = _uiState.asStateFlow()

    init {
        loadUserInfo()
        loadMedia()
    }

    private fun loadUserInfo() {
        viewModelScope.launch {
            authRepository.getCurrentUser().collect { user ->
                _uiState.update { it.copy(isAdmin = user?.isAdmin ?: false) }
            }
        }
    }

    fun loadMedia() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            mediaRepository.getMedia(mediaId).onSuccess { media ->
                _uiState.update { it.copy(isLoading = false, media = media) }
            }.onFailure { error ->
                _uiState.update { it.copy(isLoading = false, error = error.message) }
            }

            // Load progress
            mediaRepository.getProgress(mediaId).onSuccess { progress ->
                _uiState.update { it.copy(
                    progress = progress,
                    isFavorite = progress?.isFavorite ?: false,
                    isWatched = progress?.completed ?: false,
                ) }
            }

            // Load media files for tech specs
            mediaRepository.getMediaFiles(mediaId).onSuccess { files ->
                val primary = files.find { it.isPrimary } ?: files.firstOrNull()
                _uiState.update { it.copy(mediaFiles = files, primaryFile = primary) }
            }

            // Load subtitle tracks from playback info
            mediaRepository.getPlaybackInfo(mediaId, com.velox.app.data.model.PlaybackInfoRequest()).onSuccess { info ->
                _uiState.update { it.copy(
                    subtitleTracks = info.subtitleTracks.filter { track -> !track.isImage },
                ) }
            }

            // Load cinema/trailers
            loadCinema()
        }
    }

    private suspend fun loadCinema() {
        try {
            val response = veloxApi.getMediaCinema(mediaId)
            if (response.isSuccessful) {
                response.body()?.let { wrapper ->
                    _uiState.update { it.copy(cinema = wrapper.data) }
                }
            }
        } catch (e: Exception) {
            // Cinema is optional, ignore errors
        }
    }

    fun getTrailers(): List<TrailerDto> {
        return _uiState.value.cinema?.trailers ?: emptyList()
    }

    fun toggleFavorite() {
        viewModelScope.launch {
            mediaRepository.toggleFavorite(mediaId).onSuccess { isFavorite ->
                _uiState.update { it.copy(isFavorite = isFavorite) }
            }
        }
    }

    fun toggleWatched() {
        viewModelScope.launch {
            val currentState = _uiState.value
            val newWatched = !currentState.isWatched
            val position = if (newWatched) {
                (currentState.primaryFile?.duration?.toFloat() ?: 0f)
            } else {
                0f
            }
            mediaRepository.updateProgress(mediaId, position, newWatched).onSuccess {
                _uiState.update { it.copy(isWatched = newWatched) }
            }
        }
    }

    fun selectSubtitleLanguage(language: String?) {
        _uiState.update { it.copy(selectedSubtitleLanguage = language) }
    }

    fun copyStreamUrl(onResult: (String?) -> Unit) {
        viewModelScope.launch {
            mediaRepository.getStreamUrl(mediaId).onSuccess { url ->
                onResult(url)
            }.onFailure {
                onResult(null)
            }
        }
    }

    fun refreshMetadata() {
        viewModelScope.launch {
            try {
                veloxApi.refreshMediaMetadata(mediaId)
                loadMedia()
            } catch (_: Exception) { }
        }
    }

    fun refresh() {
        loadMedia()
    }
}
