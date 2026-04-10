package com.velox.app.presentation.ui.screens.splash

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.AuthManager
import com.velox.app.data.api.VeloxApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed class SplashState {
    data object Loading : SplashState()
    data object GoToHome : SplashState()
    data object GoToLogin : SplashState()
}

@HiltViewModel
class SplashViewModel @Inject constructor(
    private val authManager: AuthManager,
    private val api: VeloxApi
) : ViewModel() {

    private val _state = MutableStateFlow<SplashState>(SplashState.Loading)
    val state: StateFlow<SplashState> = _state.asStateFlow()

    init {
        checkSession()
    }

    private fun checkSession() {
        viewModelScope.launch {
            try {
                // Check if we have token
                val isLoggedIn = authManager.isLoggedIn.first()
                if (!isLoggedIn) {
                    delay(500) // Small delay for splash visibility
                    _state.value = SplashState.GoToLogin
                    return@launch
                }

                // Verify server connection and token validity by fetching profile
                val profile = api.getProfile()
                if (profile.isSuccessful) {
                    delay(500) // Small delay for splash visibility
                    _state.value = SplashState.GoToHome
                } else {
                    authManager.clearAuth()
                    _state.value = SplashState.GoToLogin
                }
            } catch (e: Exception) {
                // Network error or other failure
                _state.value = SplashState.GoToLogin
            }
        }
    }
}
