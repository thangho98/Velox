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
    val searchQuery: String = "",
    val error: String? = null,
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
            _uiState.update { it.copy(isLoading = true, error = null) }

            mediaRepository.getMediaList(
                type = "movie",
                genre = _uiState.value.selectedGenre,
                year = _uiState.value.selectedYear,
                minRating = _uiState.value.selectedRating,
                sort = _uiState.value.sortBy,
                search = _uiState.value.searchQuery.takeIf { it.isNotBlank() },
                limit = 100,
            ).onSuccess { movies ->
                _uiState.update { it.copy(isLoading = false, movies = movies) }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isLoading = false, error = error.message)
                }
            }
        }
    }

    fun setSearchQuery(query: String) {
        _uiState.update { it.copy(searchQuery = query) }
    }

    fun search() {
        loadMovies()
    }

    fun setGenre(genre: String?) {
        _uiState.update { it.copy(selectedGenre = genre) }
        loadMovies()
    }

    fun setYear(year: String?) {
        _uiState.update { it.copy(selectedYear = year) }
        loadMovies()
    }

    fun setRating(rating: Float?) {
        _uiState.update { it.copy(selectedRating = rating) }
        loadMovies()
    }

    fun setSortBy(sort: String) {
        _uiState.update { it.copy(sortBy = sort) }
        loadMovies()
    }

    fun clearFilters() {
        _uiState.update {
            it.copy(
                selectedGenre = null,
                selectedYear = null,
                selectedRating = null,
                searchQuery = "",
            )
        }
        loadMovies()
    }

    fun refresh() {
        loadMovies()
    }
}
