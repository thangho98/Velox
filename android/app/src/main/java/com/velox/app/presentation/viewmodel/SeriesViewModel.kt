package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.domain.model.SeriesItem
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SeriesUiState(
    val isLoading: Boolean = true,
    val series: List<SeriesItem> = emptyList(),
    val genres: List<String> = emptyList(),
    val selectedGenre: String? = null,
    val selectedYear: String? = null,
    val sortBy: String = "newest",
    val startChar: String? = null,
    val alphabetCounts: List<com.velox.app.domain.model.AlphabetCount> = emptyList(),
    val searchQuery: String = "",
    val error: String? = null,
    val isLoadingMore: Boolean = false,
    val hasReachedMax: Boolean = false,
    val offset: Int = 0,
    val limit: Int = 100,
)

@HiltViewModel
class SeriesViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(SeriesUiState())
    val uiState: StateFlow<SeriesUiState> = _uiState.asStateFlow()

    init {
        loadGenres()
        loadSeries()
        loadAlphabet()
    }

    private fun loadGenres() {
        viewModelScope.launch {
            mediaRepository.getGenres("series").onSuccess { genres ->
                _uiState.update {
                    it.copy(genres = genres.map { g -> g.name })
                }
            }
        }
    }

    fun loadSeries() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, offset = 0, hasReachedMax = false) }

            mediaRepository.getSeriesList(
                genre = _uiState.value.selectedGenre,
                year = _uiState.value.selectedYear,
                sort = _uiState.value.sortBy,
                search = _uiState.value.searchQuery.takeIf { it.isNotBlank() },
                limit = _uiState.value.limit,
                offset = 0,
                startChar = _uiState.value.startChar,
            ).onSuccess { items ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        series = items,
                        offset = items.size,
                        hasReachedMax = items.size < it.limit
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isLoading = false, error = error.message)
                }
            }
        }
    }

    private fun loadAlphabet() {
        viewModelScope.launch {
            mediaRepository.getSeriesAlphabet(
                genre = _uiState.value.selectedGenre,
                year = _uiState.value.selectedYear,
            ).onSuccess { counts ->
                _uiState.update { it.copy(alphabetCounts = counts) }
            }
        }
    }

    fun loadMore() {
        val state = _uiState.value
        if (state.isLoading || state.isLoadingMore || state.hasReachedMax) return

        viewModelScope.launch {
            _uiState.update { it.copy(isLoadingMore = true) }

            mediaRepository.getSeriesList(
                genre = state.selectedGenre,
                year = state.selectedYear,
                sort = state.sortBy,
                search = state.searchQuery.takeIf { it.isNotBlank() },
                limit = state.limit,
                offset = state.offset,
                startChar = state.startChar,
            ).onSuccess { newItems ->
                _uiState.update {
                    it.copy(
                        isLoadingMore = false,
                        series = it.series + newItems,
                        offset = it.offset + newItems.size,
                        hasReachedMax = newItems.size < it.limit
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isLoadingMore = false, error = error.message)
                }
            }
        }
    }

    fun setSearchQuery(query: String) {
        _uiState.update { it.copy(searchQuery = query) }
        // If the query is cleared, reset the list by reloading
        if (query.isBlank()) {
            loadSeries()
        }
    }

    fun search() {
        loadSeries()
    }

    fun setGenre(genre: String?) {
        _uiState.update { it.copy(selectedGenre = genre, startChar = null) }
        loadAlphabet()
        loadSeries()
    }

    fun setYear(year: String?) {
        _uiState.update { it.copy(selectedYear = year, startChar = null) }
        loadAlphabet()
        loadSeries()
    }

    fun setSortBy(sort: String) {
        _uiState.update { it.copy(sortBy = sort, startChar = null) }
        loadSeries()
    }

    fun setStartChar(char: String?) {
        _uiState.update { it.copy(startChar = char) }
        loadSeries()
    }

    fun clearFilters() {
        _uiState.update {
            it.copy(
                selectedGenre = null,
                selectedYear = null,
                searchQuery = "",
                startChar = null,
            )
        }
        loadAlphabet()
        loadSeries()
    }

    fun refresh() {
        loadSeries()
    }
}
