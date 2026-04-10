package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.ServerPrefsManager
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.domain.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.withContext
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed class LoginUiState {
    data object Idle : LoginUiState()
    data object Loading : LoginUiState()
    data object Success : LoginUiState()
    data class Error(val message: String) : LoginUiState()
}

enum class ConnectionStatus {
    UNKNOWN,
    CHECKING,
    CONNECTED,
    FAILED,
}

data class ConnectionInfo(
    val status: ConnectionStatus = ConnectionStatus.UNKNOWN,
    val latencyMs: Int? = null,
    val errorMessage: String? = null,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val serverPrefsManager: ServerPrefsManager,
    private val apiProvider: VeloxApiProvider,
) : ViewModel() {

    private val _uiState = MutableStateFlow<LoginUiState>(LoginUiState.Idle)
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    private val _connectionInfo = MutableStateFlow(ConnectionInfo())
    val connectionInfo: StateFlow<ConnectionInfo> = _connectionInfo.asStateFlow()

    fun getSavedServerUrl(): String {
        return serverPrefsManager.getServerUrlSync()
    }

    suspend fun saveServerUrl(url: String): String {
        return serverPrefsManager.setServerUrl(url)
    }

    suspend fun verifyServerUrl(url: String): Boolean = withContext(Dispatchers.IO) {
        try {
            var urlWithProtocol = url.trim()
            if (!urlWithProtocol.startsWith("http://") && !urlWithProtocol.startsWith("https://")) {
                urlWithProtocol = "http://$urlWithProtocol"
            }
            serverPrefsManager.setServerUrl(urlWithProtocol)
            // Brief delay to ensure DataStore persists
            delay(100)
            true
        } catch (e: Exception) {
            false
        }
    }

    /**
     * Check server connection by calling /api/health
     */
    fun checkServerConnection() {
        viewModelScope.launch {
            _connectionInfo.value = ConnectionInfo(status = ConnectionStatus.CHECKING)

            try {
                val startTime = System.currentTimeMillis()
                val response = apiProvider.getApi().healthCheck()

                if (response.isSuccessful) {
                    val latencyMs = (System.currentTimeMillis() - startTime).toInt()
                    _connectionInfo.value = ConnectionInfo(
                        status = ConnectionStatus.CONNECTED,
                        latencyMs = latencyMs,
                    )
                } else {
                    _connectionInfo.value = ConnectionInfo(
                        status = ConnectionStatus.FAILED,
                        errorMessage = "Server returned HTTP ${response.code()}",
                    )
                }
            } catch (e: Exception) {
                val errorMessage = when {
                    e.message?.contains("Unable to resolve host") == true ->
                        "Cannot reach server. Check URL and network."
                    e.message?.contains("timeout") == true ->
                        "Connection timed out"
                    else ->
                        e.message ?: "Connection failed"
                }
                _connectionInfo.value = ConnectionInfo(
                    status = ConnectionStatus.FAILED,
                    errorMessage = errorMessage,
                )
            }
        }
    }

    /**
     * Clear connection info when user changes server URL
     */
    fun clearConnectionStatus() {
        _connectionInfo.value = ConnectionInfo()
    }

    fun login(username: String, password: String) {
        if (username.isBlank() || password.isBlank()) {
            _uiState.value = LoginUiState.Error("Please enter username and password")
            return
        }

        viewModelScope.launch {
            _uiState.value = LoginUiState.Loading
            authRepository.login(username, password)
                .onSuccess {
                    _uiState.value = LoginUiState.Success
                }
                .onFailure { error ->
                    // Check if it's a network error
                    val message = error.message ?: "Login failed. Please try again."
                    if (message.contains("Network", ignoreCase = true) ||
                        message.contains("fetch", ignoreCase = true) ||
                        message.contains("connect", ignoreCase = true)) {
                        _connectionInfo.value = ConnectionInfo(
                            status = ConnectionStatus.FAILED,
                            errorMessage = "Lost connection to server",
                        )
                    }
                    _uiState.value = LoginUiState.Error(message)
                }
        }
    }

    fun clearError() {
        if (_uiState.value is LoginUiState.Error) {
            _uiState.value = LoginUiState.Idle
        }
    }
}