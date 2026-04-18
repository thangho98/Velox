package com.velox.app.presentation.viewmodel

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.VeloxApi
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.CinemaDto
import com.velox.app.data.model.TrailerDto
import com.velox.app.domain.model.Episode
import com.velox.app.domain.model.Season
import com.velox.app.domain.model.SeriesDetail
import com.velox.app.domain.model.WatchProgress
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SeriesDetailUiState(
    val isLoading: Boolean = true,
    val series: SeriesDetail? = null,
    val selectedSeason: Season? = null,
    val episodes: List<Episode> = emptyList(),
    val episodeProgress: Map<Int, WatchProgress> = emptyMap(),
    val areEpisodesLoading: Boolean = false,
    val isFavorite: Boolean = false,
    val cinema: CinemaDto? = null,
    val error: String? = null,
    val isAdmin: Boolean = false,
)

@HiltViewModel
class SeriesDetailViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
    private val authRepository: com.velox.app.domain.repository.AuthRepository,
    private val apiProvider: VeloxApiProvider,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val veloxApi: VeloxApi
        get() = apiProvider.getApi()

    val seriesId: Int = checkNotNull(savedStateHandle["seriesId"])

    private val _uiState = MutableStateFlow(SeriesDetailUiState())
    val uiState: StateFlow<SeriesDetailUiState> = _uiState.asStateFlow()

    init {
        loadUserInfo()
        loadSeries()
    }

    private fun loadUserInfo() {
        viewModelScope.launch {
            authRepository.getCurrentUser().collect { user ->
                _uiState.update { it.copy(isAdmin = user?.isAdmin ?: false) }
            }
        }
    }

    fun loadSeries() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            mediaRepository.getSeries(seriesId).onSuccess { series ->
                _uiState.update { it.copy(isLoading = false, series = series) }
                // Auto-select first season
                if (series.seasons.isNotEmpty()) {
                    selectSeason(series.seasons.first())
                }
            }.onFailure { error ->
                _uiState.update { it.copy(isLoading = false, error = error.message) }
            }

            // Load cinema/trailers
            loadCinema()
        }
    }

    private suspend fun loadCinema() {
        try {
            val response = veloxApi.getSeriesCinema(seriesId)
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

    fun selectSeason(season: Season) {
        viewModelScope.launch {
            _uiState.update { it.copy(selectedSeason = season, areEpisodesLoading = true, episodes = emptyList(), episodeProgress = emptyMap()) }

            mediaRepository.getEpisodes(seriesId, season.id).onSuccess { episodes ->
                _uiState.update { it.copy(areEpisodesLoading = false, episodes = episodes) }
                // Fetch progress for all loaded episodes concurrently
                episodes.forEach { ep ->
                    launch {
                        mediaRepository.getProgress(ep.mediaId).onSuccess { progress ->
                            if (progress != null) {
                                _uiState.update { state ->
                                    state.copy(episodeProgress = state.episodeProgress + (ep.mediaId to progress))
                                }
                            }
                        }
                    }
                }
            }.onFailure { error ->
                _uiState.update { it.copy(areEpisodesLoading = false, error = error.message) }
            }
        }
    }

    fun toggleFavorite() {
        viewModelScope.launch {
            // Series don't have direct favorite toggle - we use media items
            // This would need episode-level or series-level favorite
        }
    }

    fun refresh() {
        loadSeries()
    }

    fun autoDownloadSubtitle(mediaId: Int, onResult: (Boolean) -> Unit) {
        viewModelScope.launch {
            mediaRepository.autoDownloadSubtitle(mediaId)
                .onSuccess {
                    onResult(true)
                }
                .onFailure {
                    onResult(false)
                }
        }
    }
}
