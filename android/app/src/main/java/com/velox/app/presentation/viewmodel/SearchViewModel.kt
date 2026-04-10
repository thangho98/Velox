package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.local.RecentSearchesDataStore
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.model.SeriesItem
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SearchUiState(
    val query: String = "",
    val typeFilter: String = "", // "", "movie", "series"
    val isSearching: Boolean = false,
    val movies: List<MediaItem> = emptyList(),
    val series: List<SeriesItem> = emptyList(),
    val recentSearches: List<String> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class SearchViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
    private val recentSearchesDataStore: RecentSearchesDataStore,
) : ViewModel() {

    private val _uiState = MutableStateFlow(SearchUiState())
    val uiState: StateFlow<SearchUiState> = _uiState.asStateFlow()

    private var searchJob: Job? = null

    init {
        // Observe recent searches
        viewModelScope.launch {
            recentSearchesDataStore.recentSearches.collect { searches ->
                _uiState.update { it.copy(recentSearches = searches) }
            }
        }
    }

    fun setQuery(query: String) {
        _uiState.update { it.copy(query = query) }

        // Debounced search
        searchJob?.cancel()
        if (query.isBlank()) {
            _uiState.update { it.copy(movies = emptyList(), series = emptyList(), isSearching = false) }
            return
        }

        searchJob = viewModelScope.launch {
            delay(300) // Debounce
            search()
        }
    }

    fun search() {
        val query = _uiState.value.query
        if (query.isBlank()) return

        viewModelScope.launch {
            _uiState.update { it.copy(isSearching = true, error = null) }

            mediaRepository.search(query, limit = 20).onSuccess { (movies, series) ->
                // Save to recent searches
                recentSearchesDataStore.addSearch(query)
                _uiState.update {
                    it.copy(
                        isSearching = false,
                        movies = movies,
                        series = series,
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(
                        isSearching = false,
                        error = error.message,
                    )
                }
            }
        }
    }

    fun clearSearch() {
        _uiState.update {
            it.copy(
                query = "",
                movies = emptyList(),
                series = emptyList(),
                isSearching = false,
            )
        }
    }

    fun removeRecentSearch(query: String) {
        viewModelScope.launch {
            recentSearchesDataStore.removeSearch(query)
        }
    }

    fun setTypeFilter(type: String) {
        _uiState.update { it.copy(typeFilter = type) }
    }

    fun clearAllRecentSearches() {
        viewModelScope.launch {
            recentSearchesDataStore.clearAll()
        }
    }
}
