package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.domain.model.Genre
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MoviesUiState(
    val isLoading: Boolean = true,
    val movies: List<MediaItem> = emptyList(),
    val genres: List<String> = emptyList(),
    val selectedGenre: String? = null,
    val selectedYear: String? = null,
    val selectedRating: Float? = null,
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
class MoviesViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(MoviesUiState())
    val uiState: StateFlow<MoviesUiState> = _uiState.asStateFlow()

    init {
        loadGenres()
        loadMovies()
        loadAlphabet()
    }

    private fun loadGenres() {
        viewModelScope.launch {
            mediaRepository.getGenres("movie").onSuccess { genres ->
                _uiState.update {
                    it.copy(genres = genres.map { g -> g.name })
                }
            }
        }
    }

    fun loadMovies() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, offset = 0, hasReachedMax = false) }

            mediaRepository.getMediaList(
                type = "movie",
                genre = _uiState.value.selectedGenre,
                year = _uiState.value.selectedYear,
                minRating = _uiState.value.selectedRating,
                sort = _uiState.value.sortBy,
                search = _uiState.value.searchQuery.takeIf { it.isNotBlank() },
                limit = _uiState.value.limit,
                offset = 0,
                startChar = _uiState.value.startChar,
            ).onSuccess { movies ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        movies = movies,
                        offset = movies.size,
                        hasReachedMax = movies.size < it.limit
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
            mediaRepository.getMediaAlphabet(
                type = "movie",
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

            mediaRepository.getMediaList(
                type = "movie",
                genre = state.selectedGenre,
                year = state.selectedYear,
                minRating = state.selectedRating,
                sort = state.sortBy,
                search = state.searchQuery.takeIf { it.isNotBlank() },
                limit = state.limit,
                offset = state.offset,
                startChar = state.startChar,
            ).onSuccess { newMovies ->
                _uiState.update {
                    it.copy(
                        isLoadingMore = false,
                        movies = it.movies + newMovies,
                        offset = it.offset + newMovies.size,
                        hasReachedMax = newMovies.size < it.limit
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
            loadMovies()
        }
    }

    fun search() {
        loadMovies()
    }

    fun setGenre(genre: String?) {
        _uiState.update { it.copy(selectedGenre = genre, startChar = null) }
        loadAlphabet()
        loadMovies()
    }

    fun setYear(year: String?) {
        _uiState.update { it.copy(selectedYear = year, startChar = null) }
        loadAlphabet()
        loadMovies()
    }

    fun setRating(rating: Float?) {
        _uiState.update { it.copy(selectedRating = rating) }
        loadMovies()
    }

    fun setSortBy(sort: String) {
        _uiState.update { it.copy(sortBy = sort, startChar = null) }
        loadMovies()
    }

    fun setStartChar(char: String?) {
        _uiState.update { it.copy(startChar = char) }
        loadMovies()
    }

    fun clearFilters() {
        _uiState.update {
            it.copy(
                selectedGenre = null,
                selectedYear = null,
                selectedRating = null,
                searchQuery = "",
                startChar = null,
            )
        }
        loadAlphabet()
        loadMovies()
    }

    fun refresh() {
        loadMovies()
    }
}
