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
    val searchQuery: String = "",
    val error: String? = null,
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
            _uiState.update { it.copy(isLoading = true, error = null) }

            mediaRepository.getSeriesList(
                genre = _uiState.value.selectedGenre,
                year = _uiState.value.selectedYear,
                sort = _uiState.value.sortBy,
                search = _uiState.value.searchQuery.takeIf { it.isNotBlank() },
                limit = 100,
            ).onSuccess { items ->
                _uiState.update { it.copy(isLoading = false, series = items) }
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
        loadSeries()
    }

    fun setGenre(genre: String?) {
        _uiState.update { it.copy(selectedGenre = genre) }
        loadSeries()
    }

    fun setYear(year: String?) {
        _uiState.update { it.copy(selectedYear = year) }
        loadSeries()
    }

    fun setSortBy(sort: String) {
        _uiState.update { it.copy(sortBy = sort) }
        loadSeries()
    }

    fun clearFilters() {
        _uiState.update {
            it.copy(
                selectedGenre = null,
                selectedYear = null,
                searchQuery = "",
            )
        }
        loadSeries()
    }

    fun refresh() {
        loadSeries()
    }
}
