package com.velox.app.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.VeloxApi
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.NotificationDto
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class NotificationsUiState(
    val isLoading: Boolean = false,
    val notifications: List<NotificationDto> = emptyList(),
    val unreadCount: Int = 0,
    val error: String? = null,
)

@HiltViewModel
class NotificationsViewModel @Inject constructor(
    private val apiProvider: VeloxApiProvider,
) : ViewModel() {

    private val veloxApi: VeloxApi
        get() = apiProvider.getApi()

    private val _uiState = MutableStateFlow(NotificationsUiState())
    val uiState: StateFlow<NotificationsUiState> = _uiState.asStateFlow()

    init {
        loadNotifications()
    }

    fun loadNotifications() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                val response = veloxApi.getNotifications()
                if (response.isSuccessful) {
                    val wrapper = response.body()?.data
                    val notifications = wrapper?.notifications ?: emptyList()
                    val unreadCount = wrapper?.unreadCount ?: 0
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            notifications = notifications,
                            unreadCount = unreadCount,
                        )
                    }
                } else {
                    _uiState.update { it.copy(isLoading = false, error = "Failed to load notifications") }
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun markAsRead(notificationId: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.markNotificationsRead(
                    com.velox.app.data.model.MarkReadRequest(listOf(notificationId))
                )
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        state.copy(
                            notifications = state.notifications.map {
                                if (it.id == notificationId) it.copy(read = true) else it
                            },
                            unreadCount = state.unreadCount - 1,
                        )
                    }
                }
            } catch (ignored: Exception) {
                // Ignore errors for mark as read
            }
        }
    }

    fun markAllAsRead() {
        viewModelScope.launch {
            try {
                val response = veloxApi.markAllNotificationsRead()
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        state.copy(
                            notifications = state.notifications.map { it.copy(read = true) },
                            unreadCount = 0,
                        )
                    }
                }
            } catch (ignored: Exception) {
                // Ignore errors
            }
        }
    }

    fun deleteNotification(notificationId: Int) {
        viewModelScope.launch {
            try {
                val response = veloxApi.deleteNotifications(
                    com.velox.app.data.model.DeleteNotificationsRequest(listOf(notificationId))
                )
                if (response.isSuccessful) {
                    _uiState.update { state ->
                        val notification = state.notifications.find { it.id == notificationId }
                        state.copy(
                            notifications = state.notifications.filter { it.id != notificationId },
                            unreadCount = if (notification?.read == false) {
                                state.unreadCount - 1
                            } else {
                                state.unreadCount
                            },
                        )
                    }
                }
            } catch (ignored: Exception) {
                // Ignore errors
            }
        }
    }

    fun refresh() {
        loadNotifications()
    }
}
