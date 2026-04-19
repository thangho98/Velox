package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.domain.model.ContinueWatchingItem
import com.velox.app.domain.model.Library
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.model.NextUpItem
import com.velox.app.domain.model.SeriesItem
import com.velox.app.domain.model.User
import com.velox.app.domain.repository.AuthRepository
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class HomeUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val user: User? = null,
    val continueWatching: List<ContinueWatchingItem> = emptyList(),
    val nextUp: List<NextUpItem> = emptyList(),
    val recentlyAddedMovies: List<MediaItem> = emptyList(),
    val recentlyAddedSeries: List<SeriesItem> = emptyList(),
    val libraries: List<Library> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState.asStateFlow()

    init {
        loadData()
        observeUser()
    }

    private fun observeUser() {
        viewModelScope.launch {
            authRepository.getCurrentUser().collect { user ->
                _uiState.update { it.copy(user = user) }
            }
        }
    }

    fun loadData(isRefresh: Boolean = false) {
        viewModelScope.launch {
            if (isRefresh) {
                _uiState.update { it.copy(isRefreshing = true, error = null) }
            } else {
                _uiState.update { it.copy(isLoading = true, error = null) }
            }

            try {
                // Load libraries
                mediaRepository.getLibraries()
                    .onSuccess { libs ->
                        _uiState.update { it.copy(libraries = libs) }
                    }
                    .onFailure { /* ignore */ }

                // Load continue watching
                mediaRepository.getContinueWatching(20)
                    .onSuccess { items ->
                        _uiState.update { it.copy(continueWatching = items) }
                    }
                    .onFailure { /* ignore */ }

                // Load next up
                mediaRepository.getNextUp(20)
                    .onSuccess { items ->
                        _uiState.update { it.copy(nextUp = items) }
                    }
                    .onFailure { /* ignore */ }

                // Load recently added movies
                mediaRepository.getMediaList(type = "movie", sort = "newest", limit = 20)
                    .onSuccess { items ->
                        _uiState.update { it.copy(recentlyAddedMovies = items) }
                    }
                    .onFailure { /* ignore */ }

                // Load recently added series
                // We drop sort="newest" to match Web App which defaults to alphabetical (sort_title ASC)
                mediaRepository.getSeriesList(sort = null, limit = 20)
                    .onSuccess { items ->
                        _uiState.update { it.copy(recentlyAddedSeries = items) }
                    }
                    .onFailure { /* ignore */ }

                _uiState.update { it.copy(isLoading = false, isRefreshing = false) }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        isRefreshing = false,
                        error = e.message ?: "Failed to load data",
                    )
                }
            }
        }
    }

    fun refresh() {
        loadData(isRefresh = true)
    }

    fun dismissContinueWatching(mediaId: Int) {
        viewModelScope.launch {
            mediaRepository.dismissProgress(mediaId).onSuccess {
                _uiState.update { state ->
                    state.copy(
                        continueWatching = state.continueWatching.filter { it.mediaId != mediaId },
                    )
                }
            }
        }
    }
}
