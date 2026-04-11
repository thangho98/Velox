package com.velox.app.presentation.ui.screens.settings

import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import com.velox.app.presentation.viewmodel.SettingsViewModel
import com.velox.app.presentation.viewmodel.SettingsUiState
import com.velox.app.ui.theme.VeloxTheme
import com.velox.app.ui.theme.*

@Composable
fun ProfileScreen(
    onBackClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: SettingsViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    ProfileScreenContent(
        uiState = uiState,
        onBackClick = onBackClick,
        onSettingsClick = onSettingsClick,
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreenContent(
    uiState: SettingsUiState,
    onBackClick: () -> Unit,
    onSettingsClick: () -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Profile", color = NetflixWhite, fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = NetflixBlack),
            )
        },
        containerColor = NetflixBlack,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState()),
        ) {
            // Profile header
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                // Avatar
                val screenWidth = LocalConfiguration.current.screenWidthDp
                val avatarSize = if (screenWidth < 600) 72.dp else 96.dp
                Surface(
                    modifier = Modifier.size(avatarSize),
                    shape = RoundedCornerShape(avatarSize / 2),
                    color = NetflixDark,
                ) {
                    Box(contentAlignment = Alignment.Center) {
                        Text(
                            text = uiState.displayName.takeIf { it.isNotEmpty() }?.take(1)?.uppercase() ?: "?",
                            color = NetflixWhite,
                            fontSize = if (screenWidth < 600) 28.sp else 40.sp,
                            fontWeight = FontWeight.Bold,
                        )
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                Text(
                    text = uiState.displayName,
                    color = NetflixWhite,
                    fontSize = if (screenWidth < 600) 20.sp else 24.sp,
                    fontWeight = FontWeight.Bold,
                )

                Text(
                    text = "@${uiState.username}",
                    color = NetflixLightGray,
                    fontSize = 14.sp,
                )

                Spacer(modifier = Modifier.height(8.dp))

                Surface(
                    shape = RoundedCornerShape(16.dp),
                    color = if (uiState.isAdmin) NetflixRed.copy(alpha = 0.2f) else NetflixGray,
                ) {
                    Text(
                        text = if (uiState.isAdmin) "Administrator" else "User",
                        color = if (uiState.isAdmin) NetflixRed else NetflixLightGray,
                        fontSize = 12.sp,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp),
                    )
                }
            }

            // Stats
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                ProfileStatCard(
                    icon = LucideIcons.PlayCircle,
                    value = "0",
                    label = "Watched",
                    modifier = Modifier.weight(1f),
                )
                ProfileStatCard(
                    icon = LucideIcons.Favorite,
                    value = "0",
                    label = "Favorites",
                    modifier = Modifier.weight(1f),
                )
                ProfileStatCard(
                    icon = LucideIcons.Schedule,
                    value = "0h",
                    label = "Watch Time",
                    modifier = Modifier.weight(1f),
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Menu items
            ProfileMenuItem(
                icon = LucideIcons.Person,
                title = "Edit Profile",
                subtitle = "Change your display name",
                onClick = onSettingsClick,
            )

            ProfileMenuItem(
                icon = LucideIcons.Lock,
                title = "Change Password",
                subtitle = "Update your password",
                onClick = onSettingsClick,
            )

            ProfileMenuItem(
                icon = LucideIcons.Notifications,
                title = "Notifications",
                subtitle = "Manage notification preferences",
                onClick = onSettingsClick,
            )

            ProfileMenuItem(
                icon = LucideIcons.Logout,
                title = "Sign Out",
                subtitle = "Sign out of your account",
                onClick = { /* TODO: Implement sign out */ },
                danger = true,
            )
        }
    }
}

@Composable
private fun ProfileStatCard(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    value: String,
    label: String,
    modifier: Modifier = Modifier,
) {
    Surface(
        modifier = modifier,
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = NetflixRed,
                modifier = Modifier.size(24.dp),
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = value,
                color = NetflixWhite,
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold,
            )
            Text(
                text = label,
                color = NetflixLightGray,
                fontSize = 12.sp,
            )
        }
    }
}

@Composable
private fun ProfileMenuItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    title: String,
    subtitle: String,
    onClick: () -> Unit,
    danger: Boolean = false,
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 4.dp),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
        onClick = onClick,
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = if (danger) NetflixRed else NetflixWhite,
                modifier = Modifier.size(24.dp),
            )
            Spacer(modifier = Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title,
                    color = if (danger) NetflixRed else NetflixWhite,
                    fontWeight = FontWeight.Medium,
                )
                Text(
                    text = subtitle,
                    color = NetflixLightGray,
                    fontSize = 12.sp,
                )
            }
            Icon(
                LucideIcons.ChevronLeft,
                contentDescription = null,
                tint = NetflixLightGray,
                modifier = Modifier
                    .size(20.dp)
                    .background(
                        color = NetflixGray,
                        shape = RoundedCornerShape(10.dp),
                    )
                    .padding(2.dp),
            )
        }
    }
}

@Preview(showBackground = true)
@Composable
fun ProfileScreenPreview() {
    VeloxTheme {
        ProfileScreenContent(
            uiState = SettingsUiState(
                displayName = "John Doe",
                username = "johndoe",
                isAdmin = true
            ),
            onBackClick = {},
            onSettingsClick = {},
        )
    }
}
