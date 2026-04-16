package com.velox.app.presentation.ui.screens.settings

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.velox.app.R
import com.velox.app.data.model.NotificationDto
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.viewmodel.NotificationsUiState
import com.velox.app.presentation.viewmodel.NotificationsViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.TextMuted
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun NotificationsScreen(
    onBackClick: () -> Unit,
    viewModel: NotificationsViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    NotificationsContent(
        uiState = uiState,
        onBackClick = onBackClick,
        onMarkAllAsRead = viewModel::markAllAsRead,
        onRefresh = viewModel::refresh,
        onMarkAsRead = viewModel::markAsRead,
        onDeleteNotification = viewModel::deleteNotification
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationsContent(
    uiState: NotificationsUiState,
    onBackClick: () -> Unit,
    onMarkAllAsRead: () -> Unit,
    onRefresh: () -> Unit,
    onMarkAsRead: (Int) -> Unit,
    onDeleteNotification: (Int) -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text("Notifications", color = NetflixWhite, fontWeight = FontWeight.Bold)
                        if (uiState.unreadCount > 0) {
                            Text(
                                text = "${uiState.unreadCount} unread",
                                color = NetflixLightGray,
                                fontSize = 12.sp,
                            )
                        }
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite)
                    }
                },
                actions = {
                    if (uiState.unreadCount > 0) {
                        TextButton(onClick = onMarkAllAsRead) {
                            Text("Mark all read", color = NetflixRed)
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = NetflixBlack),
            )
        },
        containerColor = NetflixBlack,
    ) { padding ->
        if (uiState.isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = NetflixRed)
            }
        } else if (uiState.error != null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = uiState.error,
                        color = MaterialTheme.colorScheme.error,
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(
                        onClick = onRefresh,
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    ) {
                        Text(stringResource(R.string.action_retry))
                    }
                }
            }
        } else if (uiState.notifications.isEmpty()) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(
                        LucideIcons.Notifications,
                        contentDescription = null,
                        tint = NetflixGray,
                        modifier = Modifier.size(64.dp),
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = "No notifications",
                        color = NetflixLightGray,
                        fontSize = 16.sp,
                    )
                }
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(uiState.notifications) { notification ->
                    NotificationCard(
                        notification = notification,
                        onMarkAsRead = { onMarkAsRead(notification.id) },
                        onDelete = { onDeleteNotification(notification.id) },
                    )
                }
            }
        }
    }
}

@Composable
private fun NotificationCard(
    notification: NotificationDto,
    onMarkAsRead: () -> Unit,
    onDelete: () -> Unit,
) {
    val notificationIcons = mapOf(
        "scan_complete" to "🔍",
        "media_added" to "🎬",
        "transcode_complete" to "✅",
        "transcode_failed" to "❌",
        "subtitle_downloaded" to "📝",
        "identify_complete" to "🆔",
        "library_watcher" to "👁️",
    )

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(if (!notification.read) androidx.compose.ui.graphics.Color(0x1FFFFFFF) else androidx.compose.ui.graphics.Color.Transparent)
                .clickable { onMarkAsRead() }
                .padding(16.dp),
            verticalAlignment = Alignment.Top,
        ) {
            Text(
                text = notificationIcons[notification.type] ?: "🔔",
                fontSize = 20.sp,
            )

            Spacer(modifier = Modifier.width(16.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = notification.title,
                    color = if (notification.read) NetflixLightGray else NetflixWhite,
                    fontWeight = if (notification.read) FontWeight.Normal else FontWeight.Medium,
                    fontSize = 14.sp
                )
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = notification.message,
                    color = NetflixLightGray,
                    fontSize = 12.sp
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = formatTimestamp(notification.createdAt),
                    color = TextMuted,
                    fontSize = 10.sp
                )
            }

            Row(verticalAlignment = Alignment.CenterVertically) {
                if (!notification.read) {
                    IconButton(onClick = onMarkAsRead) {
                        Icon(LucideIcons.Check, contentDescription = "Mark Read", tint = NetflixLightGray, modifier = Modifier.size(16.dp))
                    }
                }
                IconButton(onClick = onDelete) {
                    Icon(LucideIcons.Delete, contentDescription = "Delete", tint = NetflixLightGray, modifier = Modifier.size(16.dp))
                }
            }
        }
    }
}

private fun formatTimestamp(timestamp: String): String {
    // Simple timestamp formatting - just return as is for now
    // Could be improved with proper date formatting
    return timestamp
}

@Preview(showBackground = true)
@Composable
fun NotificationsScreenPreview() {
    VeloxTheme {
        NotificationsContent(
            uiState = NotificationsUiState(
                notifications = listOf(SampleNotification),
                unreadCount = 1,
                isLoading = false
            ),
            onBackClick = {},
            onMarkAllAsRead = {},
            onRefresh = {},
            onMarkAsRead = {},
            onDeleteNotification = {}
        )
    }
}

@Preview(showBackground = true)
@Composable
fun NotificationCardPreview() {
    VeloxTheme {
        NotificationCard(
            notification = SampleNotification,
            onMarkAsRead = {},
            onDelete = {}
        )
    }
}

private val SampleNotification = NotificationDto(
    id = 1,
    title = "New Movie Added",
    message = "Avatar: The Way of Water is now available",
    type = "new_content",
    read = false,
    createdAt = "2024-01-15T10:30:00Z"

)
