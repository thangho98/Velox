package com.velox.app.presentation.viewmodel

import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.model.*
import com.velox.app.domain.model.User
import com.velox.app.domain.repository.AuthRepository
import com.velox.app.domain.repository.SettingsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import com.velox.app.data.api.AuthManager

data class UserProfileUiState(
    val isLoading: Boolean = false,
    val user: User? = null,
    val isAdmin: Boolean = false,
    val displayName: String = "",
    val username: String = "",

    val preferences: UserPreferences = UserPreferences(),

    val currentPassword: String = "",
    val newPassword: String = "",
    val confirmPassword: String = "",
    val securityError: String? = null,
    val securitySuccess: String? = null,

    val sessions: List<Session> = emptyList(),
    val sessionRevokeError: String? = null,

    val notificationsList: List<NotificationDto> = emptyList(),
    val unreadNotificationsCount: Int = 0,

    val error: String? = null,
    val successMessage: String? = null,
)


@HiltViewModel
class UserProfileViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    internal val settingsRepository: SettingsRepository,
    private val authManager: AuthManager,
) : ViewModel() {

    private val _uiState = MutableStateFlow(UserProfileUiState())
    val uiState: StateFlow<UserProfileUiState> = _uiState.asStateFlow()

    val currentAppLanguage: StateFlow<String> = authManager.appLanguage
        .map { it ?: "en" }
        .stateIn(viewModelScope, kotlinx.coroutines.flow.SharingStarted.Eagerly, "en")

    init {
        loadUserInfo()
    }

    private fun loadUserInfo() {
        viewModelScope.launch {
            authRepository.getCurrentUser().collect { user ->
                user?.let {
                    _uiState.update { state ->
                        state.copy(
                            user = it,
                            isAdmin = it.isAdmin,
                            displayName = it.displayName,
                            username = it.username
                        )
                    }
                }
            }
        }
    }

    fun loadProfileData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            loadPreferences()
            loadSessions()
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    private suspend fun loadPreferences() {
        settingsRepository.fetchPreferences()
            .onSuccess { prefs ->
                _uiState.update { it.copy(preferences = UserPreferences(
                    subtitleLanguage = prefs.subtitleLanguage,
                    audioLanguage = prefs.audioLanguage,
                    maxStreamingQuality = prefs.maxStreamingQuality,
                    theme = prefs.theme,
                    language = prefs.language,
                )) }
            }
    }

    private suspend fun loadSessions() {
        settingsRepository.fetchSessions()
            .onSuccess { sessions ->
                _uiState.update { it.copy(sessions = sessions.map { s ->
                    Session(
                        id = s.id,
                        deviceName = s.deviceName?.takeIf { it.isNotBlank() } ?: s.ipAddress ?: "Unknown Device",
                        ipAddress = s.ipAddress ?: "Unknown IP",
                        lastActive = s.lastActiveAt,
                        isCurrent = false,
                    )
                }) }
            }
    }

    fun updateDisplayName(name: String) {
        _uiState.update { it.copy(displayName = name) }
    }

    fun saveProfile() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            settingsRepository.saveProfile(mapOf("display_name" to _uiState.value.displayName))
                .onSuccess {
                    authRepository.getMe() // Refresh the local user state
                    _uiState.update { it.copy(successMessage = "Profile updated successfully") }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update profile: ${error.message}") }
                }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun updatePreference(field: String, value: String) {
        _uiState.update { state ->
            val p = state.preferences
            state.copy(
                preferences = when (field) {
                    "subtitleLanguage" -> p.copy(subtitleLanguage = value)
                    "audioLanguage" -> p.copy(audioLanguage = value)
                    "maxQuality" -> p.copy(maxStreamingQuality = value)
                    "theme" -> p.copy(theme = value)
                    "language" -> p.copy(language = value)
                    else -> p
                }
            )
        }
    }

    fun savePreferences() {
        val user = _uiState.value.user ?: return
        val prefs = _uiState.value.preferences

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, successMessage = null) }
            settingsRepository.savePreferences(
                UpdatePreferencesRequest(
                    userId = user.id,
                    subtitleLanguage = prefs.subtitleLanguage,
                    audioLanguage = prefs.audioLanguage,
                    maxStreamingQuality = prefs.maxStreamingQuality,
                    theme = prefs.theme,
                    language = prefs.language,
                )
            )
                .onSuccess {
                    prefs.language?.let { lang ->
                        authManager.saveAppLanguage(if (lang.isBlank()) "en" else lang)
                    }
                    _uiState.update { it.copy(successMessage = "Preferences updated successfully") }
                }
                .onError { error ->
                    _uiState.update { it.copy(error = "Failed to update preferences: ${error.message}") }
                }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun updateCurrentPassword(password: String) {
        _uiState.update { it.copy(currentPassword = password) }
    }

    fun updateNewPassword(password: String) {
        _uiState.update { it.copy(newPassword = password) }
    }

    fun updateConfirmPassword(password: String) {
        _uiState.update { it.copy(confirmPassword = password) }
    }

    fun changePassword() {
        val state = _uiState.value
        if (state.newPassword != state.confirmPassword) {
            _uiState.update { it.copy(securityError = "Passwords do not match") }
            return
        }
        if (state.newPassword.length < 8) {
            _uiState.update { it.copy(securityError = "Password must be at least 8 characters") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, securityError = null) }
            settingsRepository.changePasswordSafe(ChangePasswordRequest(
                oldPassword = state.currentPassword,
                newPassword = state.newPassword,
            ))
                .onSuccess {
                    _uiState.update {
                        it.copy(
                            securitySuccess = "Password changed successfully",
                            currentPassword = "",
                            newPassword = "",
                            confirmPassword = "",
                        )
                    }
                }
                .onError { error ->
                    _uiState.update { it.copy(securityError = "Failed to change password: ${error.message}") }
                }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun revokeSession(sessionId: Int) {
        viewModelScope.launch {
            settingsRepository.revokeSessionSafe(sessionId)
                .onSuccess {
                    _uiState.update { state ->
                        state.copy(sessions = state.sessions.filter { it.id != sessionId })
                    }
                }
                .onError {
                    _uiState.update { it.copy(sessionRevokeError = "Failed to revoke session") }
                }
        }
    }

    fun clearMessages() {
        _uiState.update {
            it.copy(
                error = null,
                successMessage = null,
                securityError = null,
                securitySuccess = null,
                sessionRevokeError = null,
            )
        }
    }
}
