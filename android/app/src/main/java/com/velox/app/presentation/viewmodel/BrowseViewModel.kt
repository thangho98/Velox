package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.domain.model.BrowseItem
import com.velox.app.domain.model.Library
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class BrowseUiState(
    val isLoading: Boolean = true,
    val libraries: List<Library> = emptyList(),
    val currentPath: List<String> = emptyList(),
    val items: List<BrowseItem> = emptyList(),
    val selectedLibraryId: Int? = null,
    val error: String? = null,
)

@HiltViewModel
class BrowseViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(BrowseUiState())
    val uiState: StateFlow<BrowseUiState> = _uiState.asStateFlow()

    init {
        loadLibraries()
    }

    private fun loadLibraries() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            mediaRepository.getLibraries().onSuccess { libraries ->
                _uiState.update { it.copy(isLoading = false, libraries = libraries) }
                // Auto-select first library
                if (libraries.isNotEmpty()) {
                    selectLibrary(libraries.first().id)
                }
            }.onFailure { error ->
                _uiState.update { it.copy(isLoading = false, error = error.message) }
            }
        }
    }

    fun selectLibrary(libraryId: Int) {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isLoading = true,
                    selectedLibraryId = libraryId,
                    currentPath = emptyList(),
                    error = null,
                )
            }

            mediaRepository.browse(libraryId = libraryId, path = null)
                .onSuccess { items ->
                    _uiState.update { it.copy(isLoading = false, items = items) }
                }
                .onFailure { error ->
                    _uiState.update { it.copy(isLoading = false, error = error.message) }
                }
        }
    }

    fun navigateToPath(path: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            val pathParts = path.split("/").filter { it.isNotEmpty() }
            _uiState.update { it.copy(currentPath = pathParts) }

            mediaRepository.browse(
                libraryId = _uiState.value.selectedLibraryId,
                path = path,
            ).onSuccess { items ->
                _uiState.update { it.copy(isLoading = false, items = items) }
            }.onFailure { error ->
                _uiState.update { it.copy(isLoading = false, error = error.message) }
            }
        }
    }

    fun navigateBack(): Boolean {
        val currentPath = _uiState.value.currentPath
        if (currentPath.isEmpty()) {
            return false
        }
        val parentPath = currentPath.dropLast(1).joinToString("/")
        navigateToPath(parentPath)
        return true
    }

    fun refresh() {
        val currentPath = _uiState.value.currentPath.joinToString("/")
        if (currentPath.isEmpty()) {
            _uiState.value.selectedLibraryId?.let { selectLibrary(it) }
        } else {
            navigateToPath(currentPath)
        }
    }
}
