package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class FavoritesUiState(
    val isLoading: Boolean = true,
    val favorites: List<MediaItem> = emptyList(),
    val error: String? = null,
    val isLoadingMore: Boolean = false,
    val hasReachedMax: Boolean = false,
    val offset: Int = 0,
    val limit: Int = 100,
)

@HiltViewModel
class FavoritesViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(FavoritesUiState())
    val uiState: StateFlow<FavoritesUiState> = _uiState.asStateFlow()

    init {
        loadFavorites()
    }

    fun loadFavorites() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, offset = 0, hasReachedMax = false) }

            mediaRepository.getFavorites(limit = _uiState.value.limit, offset = 0).onSuccess { items ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        favorites = items,
                        offset = items.size,
                        hasReachedMax = items.size < it.limit
                    )
                }
            }.onFailure { error ->
                _uiState.update { it.copy(isLoading = false, error = error.message) }
            }
        }
    }

    fun loadMore() {
        val state = _uiState.value
        if (state.isLoading || state.isLoadingMore || state.hasReachedMax) return

        viewModelScope.launch {
            _uiState.update { it.copy(isLoadingMore = true) }

            mediaRepository.getFavorites(limit = state.limit, offset = state.offset).onSuccess { newItems ->
                _uiState.update {
                    it.copy(
                        isLoadingMore = false,
                        favorites = it.favorites + newItems,
                        offset = it.offset + newItems.size,
                        hasReachedMax = newItems.size < it.limit
                    )
                }
            }.onFailure { error ->
                _uiState.update { it.copy(isLoadingMore = false, error = error.message) }
            }
        }
    }

    fun toggleFavorite(mediaId: Int) {
        viewModelScope.launch {
            mediaRepository.toggleFavorite(mediaId).onSuccess { isFavorite ->
                if (!isFavorite) {
                    // Remove from list
                    _uiState.update { state ->
                        state.copy(favorites = state.favorites.filter { it.id != mediaId })
                    }
                }
            }
        }
    }

    fun refresh() {
        loadFavorites()
    }
}
