package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.model.*
import com.velox.app.domain.model.Library
import com.velox.app.domain.repository.MediaRepository
import com.velox.app.domain.repository.SettingsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import timber.log.Timber
import javax.inject.Inject

data class SystemAdminUiState(
    val isLoading: Boolean = false,

    // Dashboard Stats
    val serverInfo: ServerInfoDto? = null,
    val libraryStats: List<LibraryStatsDto> = emptyList(),
    val activityLogs: List<ActivityLogDto> = emptyList(),

    // Libraries
    val libraries: List<Library> = emptyList(),
    val scanningLibraryId: Int? = null,

    // Users
    val users: List<AdminUser> = emptyList(),
    val currentUserId: Int? = null,

    // Tasks
    val tasks: List<TaskDto> = emptyList(),
    val runningTaskName: String? = null,

    // Webhooks
    val webhooks: List<Webhook> = emptyList(),

    val error: String? = null,
    val successMessage: String? = null,
    // App Version
    val latestVersion: String? = null,
    val updateDownloadUrl: String? = null,
    val updateIsMandatory: Boolean = false,
    val updateStatus: String = "",
    val isCheckingUpdate: Boolean = false,
)

@HiltViewModel
class SystemAdminViewModel @Inject constructor(
    private val settingsRepository: SettingsRepository,
    private val mediaRepository: MediaRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(SystemAdminUiState())
    val uiState: StateFlow<SystemAdminUiState> = _uiState.asStateFlow()

    fun loadDashboardData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            settingsRepository.fetchServerInfo()
                .onSuccess { info -> _uiState.update { it.copy(serverInfo = info) } }

            settingsRepository.fetchLibraryStats()
                .onSuccess { stats -> _uiState.update { it.copy(libraryStats = stats) } }

            settingsRepository.fetchAdminActivity(limit = 10)
                .onSuccess { activities -> _uiState.update { it.copy(activityLogs = activities) } }

            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun loadLibraries() {
        viewModelScope.launch {
            mediaRepository.getLibraries().onSuccess { libraries ->
                _uiState.update { state -> state.copy(libraries = libraries) }
            }
        }
    }

    fun loadActivity(limit: Int = 50, offset: Int = 0) {
        viewModelScope.launch {
            settingsRepository.fetchAdminActivity(limit, offset)
                .onSuccess { logs -> _uiState.update { it.copy(activityLogs = logs) } }
                .onError { Timber.e("loadActivity: ${it.message}") }
        }
    }

    fun loadUsers() {
        viewModelScope.launch {
            // First fetch the current user to get its ID
            var currentId: Int? = null
            settingsRepository.fetchProfile().onSuccess { p -> currentId = p.id }

            settingsRepository.fetchAdminUsers()
                .onSuccess { users ->
                    _uiState.update { it.copy(
                        currentUserId = currentId,
                        users = users.map { u ->
                            AdminUser(
                                id = u.id,
                                username = u.username,
                                displayName = u.displayName,
                                isAdmin = u.isAdmin,
                                lastActive = u.lastActive,
                                createdAt = u.createdAt,
                            )
                        }
                    ) }
                }
        }
    }

    fun loadWebhooks() {
        viewModelScope.launch {
            settingsRepository.fetchWebhooks()
                .onSuccess { webhooks ->
                    _uiState.update { it.copy(webhooks = webhooks.map { w ->
                        Webhook(
                            id = w.id,
                            url = w.url,
                            events = w.events,
                            active = w.active,
                        )
                    }) }
                }
        }
    }

    fun loadTasks() {
        viewModelScope.launch {
            settingsRepository.fetchTasks()
                .onSuccess { ts -> _uiState.update { it.copy(tasks = ts) } }
        }
    }

    fun createLibrary(name: String, type: String, paths: List<String>) {
        viewModelScope.launch {
            settingsRepository.createLibrarySafe(CreateLibraryRequest(name = name, type = type, paths = paths))
                .onSuccess { loadLibraries() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to create library: ${error.message}") } }
        }
    }

    fun updateLibrary(id: Int, name: String, paths: List<String>) {
        viewModelScope.launch {
            settingsRepository.updateLibrarySafe(id, UpdateLibraryRequest(name = name, paths = paths))
                .onSuccess { loadLibraries() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to update library: ${error.message}") } }
        }
    }

    fun deleteLibrary(id: Int) {
        viewModelScope.launch {
            settingsRepository.deleteLibrarySafe(id)
                .onSuccess {
                    _uiState.update { state -> state.copy(libraries = state.libraries.filter { it.id != id }) }
                }
                .onError { error -> _uiState.update { it.copy(error = "Failed to delete library: ${error.message}") } }
        }
    }

    fun scanLibrary(id: Int, force: Boolean) {
        viewModelScope.launch {
            _uiState.update { it.copy(scanningLibraryId = id) }
            settingsRepository.scanLibrarySafe(id, force)
                .onSuccess { _uiState.update { it.copy(successMessage = "Scan started") } }
                .onError { error -> _uiState.update { it.copy(error = "Failed to scan library: ${error.message}") } }
            _uiState.update { it.copy(scanningLibraryId = null) }
        }
    }

    fun createUser(username: String, password: String, displayName: String, isAdmin: Boolean) {
        viewModelScope.launch {
            settingsRepository.createUserSafe(
                CreateUserRequest(username = username, password = password, displayName = displayName, isAdmin = isAdmin)
            )
                .onSuccess { loadUsers() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to create user: ${error.message}") } }
        }
    }

    fun updateUser(userId: Int, displayName: String?, isAdmin: Boolean?) {
        viewModelScope.launch {
            settingsRepository.updateUserSafe(userId, UpdateUserRequest(displayName = displayName, isAdmin = isAdmin))
                .onSuccess { loadUsers() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to update user: ${error.message}") } }
        }
    }

    fun deleteUser(userId: Int) {
        viewModelScope.launch {
            settingsRepository.deleteUserSafe(userId)
                .onSuccess {
                    _uiState.update { state -> state.copy(users = state.users.filter { it.id != userId }) }
                }
                .onError { error -> _uiState.update { it.copy(error = "Failed to delete user: ${error.message}") } }
        }
    }

    fun createWebhook(url: String, events: List<String>, active: Boolean) {
        viewModelScope.launch {
            settingsRepository.createWebhookSafe(CreateWebhookRequest(url = url, events = events, active = active))
                .onSuccess { loadWebhooks() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to create webhook: ${error.message}") } }
        }
    }

    fun updateWebhook(id: Int, url: String, events: List<String>, active: Boolean) {
        viewModelScope.launch {
            settingsRepository.updateWebhookSafe(id, UpdateWebhookRequest(url = url, events = events, active = active))
                .onSuccess { loadWebhooks() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to update webhook: ${error.message}") } }
        }
    }

    fun deleteWebhook(id: Int) {
        viewModelScope.launch {
            settingsRepository.deleteWebhookSafe(id)
                .onSuccess {
                    _uiState.update { state -> state.copy(webhooks = state.webhooks.filter { it.id != id }) }
                }
                .onError { error -> _uiState.update { it.copy(error = "Failed to delete webhook: ${error.message}") } }
        }
    }

    fun runTask(name: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(runningTaskName = name) }
            settingsRepository.runTaskSafe(name)
                .onSuccess { loadTasks() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to run task: ${error.message}") } }
            _uiState.update { it.copy(runningTaskName = null) }
        }
    }

    fun updateTaskInterval(name: String, interval: String) {
        viewModelScope.launch {
            settingsRepository.updateTaskIntervalSafe(name, UpdateTaskIntervalRequest(interval))
                .onSuccess { loadTasks() }
                .onError { error -> _uiState.update { it.copy(error = "Failed to update task interval: ${error.message}") } }
        }
    }

    fun clearMessages() {
        _uiState.update { it.copy(error = null, successMessage = null) }
    }
    fun checkAppUpdate() {
        viewModelScope.launch {
            _uiState.update { it.copy(isCheckingUpdate = true, updateStatus = "", error = null) }
            settingsRepository.fetchLatestAppVersion()
                .onSuccess { body ->
                    val hasUpdate = body.versionCode > com.velox.app.BuildConfig.VERSION_CODE
                    _uiState.update {
                        it.copy(
                            latestVersion = if (hasUpdate) body.versionName else null,
                            updateDownloadUrl = if (hasUpdate) body.downloadUrl else null,
                            updateIsMandatory = hasUpdate && body.isMandatory,
                            updateStatus = if (hasUpdate) "Version ${body.versionName} is available" else "You're up to date"
                        )
                    }
                }
                .onError { error -> _uiState.update { it.copy(updateStatus = "Failed to check for updates: ${error.message}") } }
            _uiState.update { it.copy(isCheckingUpdate = false) }
        }
    }

}
