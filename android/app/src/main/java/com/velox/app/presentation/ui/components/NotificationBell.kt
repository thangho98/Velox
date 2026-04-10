package com.velox.app.presentation.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.velox.app.data.api.VeloxApi
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class NotificationBellViewModel @Inject constructor(
    private val veloxApi: VeloxApi,
) : ViewModel() {
    private val _unreadCount = MutableStateFlow(0)
    val unreadCount = _unreadCount.asStateFlow()

    init {
        pollUnreadCount()
    }

    private fun pollUnreadCount() {
        viewModelScope.launch {
            while (true) {
                try {
                    val response = veloxApi.getUnreadCount()
                    if (response.isSuccessful) {
                        _unreadCount.value = response.body()?.data?.count ?: 0
                    }
                } catch (e: Exception) {
                    // Ignore
                }
                delay(30000) // Poll every 30s
            }
        }
    }
}

@Composable
fun NotificationBell(
    onClick: () -> Unit,
    viewModel: NotificationBellViewModel = hiltViewModel()
) {
    val unreadCount by viewModel.unreadCount.collectAsState()

    // Dùng wrapper Box bên ngoài IconButton để tránh bị clip bởi shape của IconButton
    Box(contentAlignment = Alignment.Center) {
        IconButton(onClick = onClick) {
            Icon(
                LucideIcons.Notifications,
                contentDescription = "Notifications",
                tint = NetflixWhite,
                modifier = Modifier.size(24.dp)
            )
        }
        if (unreadCount > 0) {
            val countText = if (unreadCount > 9) "9+" else unreadCount.toString()
            Box(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    // (48 - 24) / 2 = 12dp. 
                    // Để đưa badge về đỉnh icon thì cần x=-12, y=12.
                    // Kéo cao lên xíu và dịch phải xíu so với icon -> x = -8.dp, y = 8.dp
                    .offset(x = (-8).dp, y = 8.dp)
                    .zIndex(1f)
                    .widthIn(min = 18.dp)
                    .height(18.dp)
                    .background(NetflixRed, RoundedCornerShape(percent = 50)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = countText,
                    color = NetflixWhite,
                    fontSize = 10.sp,
                    fontWeight = FontWeight.Bold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.padding(horizontal = 4.dp).offset(y = (-0.5).dp)
                )
            }
        }
    }
}
