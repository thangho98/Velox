package com.velox.app.presentation.ui.screens.settings

import androidx.compose.foundation.background
import androidx.compose.foundation.BorderStroke
import com.velox.app.presentation.ui.components.LucideIcons
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import kotlinx.coroutines.launch
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import android.content.Intent
import android.net.Uri
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import androidx.compose.foundation.Image
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextAlign
import com.velox.app.presentation.viewmodel.*
import com.velox.app.domain.model.Library
import com.velox.app.ui.theme.TextMuted
import com.velox.app.ui.theme.*

// Tab-based navigation for Settings
enum class SettingsTab(val title: String) {
    PROFILE("Profile"),
    PREFERENCES("Preferences"),
    SECURITY("Security"),
    SESSIONS("Sessions"),
    METADATA("Metadata"),
    SUBTITLES("Subtitles"),
    PLAYBACK("Playback"),
    CINEMA("Cinema Mode"),
    PRETRANSCODE("Pre-transcode"),
    MARKERS("Skip Intro / Credits"),
    DASHBOARD("Dashboard"),
    LIBRARIES("Libraries"),
    USERS("Users"),
    ACTIVITY("Activity"),
    TASKS("Tasks"),
    WEBHOOKS("Webhooks")
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    onBackClick: () -> Unit,
    onProfileClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    viewModel: SettingsViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    SettingsContent(
        uiState = uiState,
        onBackClick = onBackClick,
        onProfileClick = onProfileClick,
        onNotificationsClick = onNotificationsClick,
        onAction = { action ->
            when (action) {
                is SettingsAction.UpdateLibrary -> viewModel.updateLibrary(action.id, action.name, action.paths)
                is SettingsAction.CreateLibrary -> viewModel.createLibrary(action.name, action.type, action.paths)
                is SettingsAction.UpdateUser -> viewModel.updateUser(action.id, action.displayName, action.isAdmin)
                is SettingsAction.CreateUser -> viewModel.createUser(action.username, action.password, action.displayName, action.isAdmin)
                is SettingsAction.UpdateWebhook -> viewModel.updateWebhook(action.id, action.url, action.events, action.active)
                is SettingsAction.CreateWebhook -> viewModel.createWebhook(action.url, action.events, action.active)
                is SettingsAction.SelectSection -> viewModel.selectSection(action.section)
                is SettingsAction.SelectTab -> viewModel.setSelectedTab(action.tab)
                is SettingsAction.DeleteLibrary -> viewModel.deleteLibrary(action.id)
                is SettingsAction.ScanLibrary -> viewModel.scanLibrary(action.id, action.force)
                is SettingsAction.DeleteUser -> viewModel.deleteUser(action.id)
                is SettingsAction.DeleteWebhook -> viewModel.deleteWebhook(action.id)
            }
        },
        viewModel = viewModel // Temporarily pass viewModel until sub-sections are refactored
    )
}

sealed class SettingsAction {
    data class UpdateLibrary(val id: Int, val name: String, val paths: List<String>) : SettingsAction()
    data class CreateLibrary(val name: String, val type: String, val paths: List<String>) : SettingsAction()
    data class UpdateUser(val id: Int, val displayName: String, val isAdmin: Boolean) : SettingsAction()
    data class CreateUser(val username: String, val password: String, val displayName: String, val isAdmin: Boolean) : SettingsAction()
    data class UpdateWebhook(val id: Int, val url: String, val events: List<String>, val active: Boolean) : SettingsAction()
    data class CreateWebhook(val url: String, val events: List<String>, val active: Boolean) : SettingsAction()
    data class SelectSection(val section: SettingsSection) : SettingsAction()
    data class SelectTab(val tab: String) : SettingsAction()
    data class DeleteLibrary(val id: Int) : SettingsAction()
    data class ScanLibrary(val id: Int, val force: Boolean = false) : SettingsAction()
    data class DeleteUser(val id: Int) : SettingsAction()
    data class DeleteWebhook(val id: Int) : SettingsAction()
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsContent(
    uiState: SettingsUiState,
    onBackClick: () -> Unit,
    onProfileClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onAction: (SettingsAction) -> Unit,
    viewModel: SettingsViewModel? = null, // Optional for sub-sections that haven't been refactored
) {
    // Dialog states
    var showAddLibraryDialog by remember { mutableStateOf(false) }
    var editingLibrary by remember { mutableStateOf<Library?>(null) }
    var showAddUserDialog by remember { mutableStateOf(false) }
    var editingUser by remember { mutableStateOf<AdminUser?>(null) }
    var showAddWebhookDialog by remember { mutableStateOf(false) }
    var editingWebhook by remember { mutableStateOf<Webhook?>(null) }
    
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(uiState.error) {
        uiState.error?.let { snackbarHostState.showSnackbar(it) }
    }
    
    LaunchedEffect(uiState.successMessage) {
        uiState.successMessage?.let { snackbarHostState.showSnackbar(it) }
    }

    // Dialogs
    if (showAddLibraryDialog || editingLibrary != null) {
        LibraryDialog(
            library = editingLibrary,
            onDismiss = {
                showAddLibraryDialog = false
                editingLibrary = null
            },
            onSave = { name, type, paths ->
                if (editingLibrary != null) {
                    onAction(SettingsAction.UpdateLibrary(editingLibrary!!.id, name, paths))
                } else {
                    onAction(SettingsAction.CreateLibrary(name, type, paths))
                }
                showAddLibraryDialog = false
                editingLibrary = null
            },
        )
    }

    if (showAddUserDialog || editingUser != null) {
        UserDialog(
            user = editingUser,
            onDismiss = {
                showAddUserDialog = false
                editingUser = null
            },
            onSave = { username, password, displayName, isAdmin ->
                if (editingUser != null) {
                    onAction(SettingsAction.UpdateUser(editingUser!!.id, displayName, isAdmin))
                } else {
                    onAction(SettingsAction.CreateUser(username, password, displayName, isAdmin))
                }
                showAddUserDialog = false
                editingUser = null
            },
        )
    }

    if (showAddWebhookDialog || editingWebhook != null) {
        WebhookDialog(
            webhook = editingWebhook,
            onDismiss = {
                showAddWebhookDialog = false
                editingWebhook = null
            },
            onSave = { url, events, active ->
                if (editingWebhook != null) {
                    onAction(SettingsAction.UpdateWebhook(editingWebhook!!.id, url, events, active))
                } else {
                    onAction(SettingsAction.CreateWebhook(url, events, active))
                }
                showAddWebhookDialog = false
                editingWebhook = null
            },
        )
    }

    Scaffold(
        snackbarHost = { SnackbarHost(hostState = snackbarHostState) },
        topBar = {
            TopAppBar(
                title = { 
                    val titleText = if (uiState.selectedTab == "menu") "Settings" else {
                        SettingsTab.values().find { it.name.lowercase() == uiState.selectedTab }?.title ?: "Settings"
                    }
                    Text(titleText, color = NetflixWhite, fontWeight = FontWeight.Bold) 
                },
                navigationIcon = {
                    IconButton(onClick = {
                        if (uiState.selectedTab == "menu") onBackClick()
                        else onAction(SettingsAction.SelectTab("menu"))
                    }) {
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
                .padding(padding),
        ) {
            // User card
            uiState.user?.let { user ->
                UserCard(
                    displayName = user.displayName,
                    username = user.username,
                    isAdmin = user.isAdmin,
                    profilePath = user.profilePath,
                    onClick = onProfileClick,
                )
            }

                        if (uiState.selectedTab == "menu") {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(bottom = 32.dp)
                ) {
                    item {
                        Text(
                            text = "WEB SETTINGS",
                            color = NetflixLightGray,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Bold,
                            modifier = Modifier.padding(start = 24.dp, top = 24.dp, bottom = 8.dp)
                        )
                        val webSettings = listOf(SettingsTab.PROFILE, SettingsTab.PREFERENCES, SettingsTab.SECURITY, SettingsTab.SESSIONS)
                        SettingsGroupCard(tabs = webSettings, onAction = onAction)
                    }

                    if (uiState.isAdmin) {
                        item {
                            Text(
                                text = "ADMIN PREFERENCES",
                                color = NetflixLightGray,
                                fontSize = 12.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 24.dp, top = 24.dp, bottom = 8.dp)
                            )
                            val adminPrefs = listOf(SettingsTab.METADATA, SettingsTab.SUBTITLES, SettingsTab.PLAYBACK, SettingsTab.CINEMA, SettingsTab.PRETRANSCODE, SettingsTab.MARKERS)
                            SettingsGroupCard(tabs = adminPrefs, onAction = onAction)
                        }

                        item {
                            Text(
                                text = "VELOX SERVER",
                                color = NetflixLightGray,
                                fontSize = 12.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 24.dp, top = 24.dp, bottom = 8.dp)
                            )
                            val serverPrefs = listOf(SettingsTab.DASHBOARD, SettingsTab.LIBRARIES, SettingsTab.USERS, SettingsTab.ACTIVITY, SettingsTab.TASKS, SettingsTab.WEBHOOKS)
                            SettingsGroupCard(tabs = serverPrefs, onAction = onAction)
                        }
                    }

                    item {
                        FeedbackCard()
                    }

                    item {
                        AppVersionCard()
                    }
                }
            } else {
            if (uiState.isLoading) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = NetflixRed)
                }
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    item {
                        val currentSection = when (uiState.selectedTab) {
                            "profile" -> SettingsSection.PROFILE
                            "preferences" -> SettingsSection.PREFERENCES
                            "security" -> SettingsSection.SECURITY
                            "sessions" -> SettingsSection.SESSIONS
                            "metadata" -> SettingsSection.METADATA
                            "subtitles" -> SettingsSection.SUBTITLES
                            "playback" -> SettingsSection.PLAYBACK
                            "cinema" -> SettingsSection.CINEMA
                            "pretranscode" -> SettingsSection.PRETRANSCODE
                            "markers" -> SettingsSection.MARKERS
                            "dashboard" -> SettingsSection.DASHBOARD
                            "libraries" -> SettingsSection.LIBRARIES
                            "users" -> SettingsSection.USERS
                            "activity" -> SettingsSection.ACTIVITY
                            "tasks" -> SettingsSection.TASKS
                            "webhooks" -> SettingsSection.WEBHOOKS
                            else -> SettingsSection.PROFILE
                        }
                        when (currentSection) {
                            SettingsSection.PROFILE -> viewModel?.let { ProfileSection(it, uiState) }
                            SettingsSection.PREFERENCES -> viewModel?.let { PreferencesSection(it, uiState) }
                            SettingsSection.SECURITY -> viewModel?.let { SecuritySection(it, uiState) }
                            SettingsSection.SESSIONS -> viewModel?.let { SessionsSection(it, uiState) }
                            SettingsSection.METADATA -> viewModel?.let { MetadataSection(it, uiState) }
                            SettingsSection.SUBTITLES -> viewModel?.let { SubtitlesSection(it, uiState) }
                            SettingsSection.PLAYBACK -> viewModel?.let { PlaybackSection(it, uiState) }
                            SettingsSection.CINEMA -> viewModel?.let { CinemaSection(it, uiState) }
                            SettingsSection.PRETRANSCODE -> viewModel?.let { PretranscodeSection(it, uiState) }
                            SettingsSection.MARKERS -> viewModel?.let { MarkersSection(it, uiState) }
                            SettingsSection.DASHBOARD -> viewModel?.let { DashboardSection(it, uiState) }
                            SettingsSection.LIBRARIES -> LibrariesSection(
                                viewModel = viewModel,
                                uiState = uiState,
                                onAddClick = { showAddLibraryDialog = true },
                                onEditClick = { editingLibrary = it },
                                onDeleteClick = { onAction(SettingsAction.DeleteLibrary(it)) },
                                onScanClick = { id, force -> onAction(SettingsAction.ScanLibrary(id, force)) },
                            )
                            SettingsSection.USERS -> UsersSection(
                                viewModel = viewModel,
                                uiState = uiState,
                                onAddClick = { showAddUserDialog = true },
                                onEditClick = { editingUser = it },
                                onDeleteClick = { onAction(SettingsAction.DeleteUser(it)) },
                            )
                            SettingsSection.ACTIVITY -> viewModel?.let { ActivitySection(it, uiState) }
                            SettingsSection.TASKS -> viewModel?.let { TasksSection(it, uiState) }
                            SettingsSection.WEBHOOKS -> WebhooksSection(
                                viewModel = viewModel,
                                uiState = uiState,
                                onAddClick = { showAddWebhookDialog = true },
                                onEditClick = { editingWebhook = it },
                                onDeleteClick = { onAction(SettingsAction.DeleteWebhook(it)) },
                            )
                        }
                    }
                }
            }
            }
        }
    }
}

@Composable
private fun UserCard(
    displayName: String,
    username: String,
    isAdmin: Boolean,
    profilePath: String?,
    onClick: () -> Unit,
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp)
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick),
        color = NetflixDark,
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Surface(
                modifier = Modifier.size(48.dp),
                shape = RoundedCornerShape(24.dp),
                color = NetflixGray,
            ) {
                if (profilePath != null) {
                    AsyncImage(
                        model = profilePath,
                        contentDescription = "Profile Picture",
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    Box(contentAlignment = Alignment.Center) {
                        Text(
                            text = displayName.take(1).uppercase(),
                            color = NetflixWhite,
                            fontWeight = FontWeight.Bold,
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.width(16.dp))
            Column {
                Text(
                    text = displayName,
                    color = NetflixWhite,
                    fontWeight = FontWeight.SemiBold,
                )
                Text(
                    text = if (isAdmin) "Administrator" else "User",
                    color = if (isAdmin) NetflixRed else NetflixLightGray,
                    fontSize = 12.sp,
                )
            }
        }
    }
}

@Composable
private fun ProfileSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Profile", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Manage your account information", color = NetflixLightGray, fontSize = 14.sp)

        OutlinedTextField(
            value = uiState.username,
            onValueChange = {},
            label = { Text("Username") },
            enabled = false,
            modifier = Modifier.fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                disabledTextColor = NetflixWhite,
                disabledBorderColor = NetflixGray,
                disabledLabelColor = NetflixLightGray,
            ),
        )
        Text("Username cannot be changed", color = NetflixLightGray, fontSize = 12.sp)

        OutlinedTextField(
            value = uiState.displayName,
            onValueChange = viewModel::updateDisplayName,
            label = { Text("Display Name") },
            modifier = Modifier.fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                focusedTextColor = NetflixWhite,
                unfocusedTextColor = NetflixWhite,
                focusedBorderColor = NetflixRed,
                unfocusedBorderColor = NetflixGray,
                focusedLabelColor = NetflixRed,
                unfocusedLabelColor = NetflixLightGray,
            ),
        )

        // Role - read-only
        OutlinedTextField(
            value = if (uiState.isAdmin) "Administrator" else "User",
            onValueChange = {},
            label = { Text("Role") },
            enabled = false,
            modifier = Modifier.fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                disabledTextColor = NetflixLightGray,
                disabledBorderColor = NetflixGray,
                disabledLabelColor = NetflixLightGray,
            ),
        )

        uiState.error?.let {
            Text(it, color = NetflixRed, fontSize = 14.sp)
        }
        
        Button(
            onClick = viewModel::saveProfile,
            modifier = Modifier.wrapContentWidth(Alignment.Start),
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(8.dp),
            enabled = !uiState.isLoading,
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = NetflixWhite,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Save Changes")
            }
        }

        uiState.successMessage?.let {
            Text(it, color = androidx.compose.ui.graphics.Color(0xFF4CAF50), fontSize = 14.sp)
        }
    }
}

@Composable
private fun PreferencesSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Preferences", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Customize your streaming experience", color = NetflixLightGray, fontSize = 14.sp)

        SettingsDropdown(
            label = "Subtitle Language",
            value = uiState.preferences.subtitleLanguage ?: "",
            options = listOf("" to "Auto", "vi" to "Tiếng Việt", "en" to "English"),
            onValueChange = { viewModel.updatePreference("subtitleLanguage", it) },
        )

        SettingsDropdown(
            label = "Audio Language",
            value = uiState.preferences.audioLanguage ?: "",
            options = listOf("" to "Auto", "vi" to "Tiếng Việt", "en" to "English"),
            onValueChange = { viewModel.updatePreference("audioLanguage", it) },
        )

        SettingsDropdown(
            label = "Max Streaming Quality",
            value = if (uiState.preferences.maxStreamingQuality.isNullOrBlank()) "original" else uiState.preferences.maxStreamingQuality,
            options = listOf("original" to "Original", "4k" to "4K (2160p)", "1080p" to "HD (1080p)", "720p" to "HD (720p)", "480p" to "SD (480p)"),
            onValueChange = { viewModel.updatePreference("maxStreamingQuality", it) },
        )

        SettingsDropdown(
            label = "Theme",
            value = if (uiState.preferences.theme.isNullOrBlank()) "system" else uiState.preferences.theme,
            options = listOf("system" to "System", "dark" to "Dark", "light" to "Light"),
            onValueChange = { viewModel.updatePreference("theme", it) },
        )

        SettingsDropdown(
            label = "Language",
            value = if (uiState.preferences.language.isNullOrBlank()) "en" else uiState.preferences.language,
            options = listOf("en" to "English", "vi" to "Tiếng Việt"),
            onValueChange = { viewModel.updatePreference("language", it) },
        )



        uiState.error?.let {
            Text(it, color = NetflixRed, fontSize = 14.sp)
        }
        
        Button(
            onClick = viewModel::savePreferences,
            modifier = Modifier.wrapContentWidth(Alignment.Start),
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(8.dp),
            enabled = !uiState.isLoading,
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = NetflixWhite,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Save Changes")
            }
        }

        uiState.successMessage?.let {
            Text(it, color = androidx.compose.ui.graphics.Color(0xFF4CAF50), fontSize = 14.sp)
        }
    }
}

@Composable
private fun SecuritySection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Security", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Change your password", color = NetflixLightGray, fontSize = 14.sp)

        OutlinedTextField(
            value = uiState.currentPassword,
            onValueChange = viewModel::updateCurrentPassword,
            label = { Text("Current Password") },
            modifier = Modifier.fillMaxWidth(),
            colors = textFieldColors(),
            shape = RoundedCornerShape(8.dp),
            visualTransformation = PasswordVisualTransformation(),
        )

        OutlinedTextField(
            value = uiState.newPassword,
            onValueChange = viewModel::updateNewPassword,
            label = { Text("New Password") },
            modifier = Modifier.fillMaxWidth(),
            colors = textFieldColors(),
            shape = RoundedCornerShape(8.dp),
            visualTransformation = PasswordVisualTransformation(),
        )

        OutlinedTextField(
            value = uiState.confirmPassword,
            onValueChange = viewModel::updateConfirmPassword,
            label = { Text("Confirm Password") },
            modifier = Modifier.fillMaxWidth(),
            colors = textFieldColors(),
            shape = RoundedCornerShape(8.dp),
            visualTransformation = PasswordVisualTransformation(),
        )

        uiState.securityError?.let {
            Text(it, color = NetflixRed, fontSize = 14.sp)
        }

        uiState.securitySuccess?.let {
            Text(it, color = androidx.compose.ui.graphics.Color(0xFF4CAF50), fontSize = 14.sp)
        }

        Button(
            onClick = viewModel::changePassword,
            modifier = Modifier.wrapContentWidth(Alignment.Start),
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(8.dp),
            enabled = !uiState.isLoading,
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = NetflixWhite,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Update Password")
            }
        }
    }
}

@Composable
private fun SessionsSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Active Sessions", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Manage devices signed into your account", color = NetflixLightGray, fontSize = 14.sp)

        if (uiState.isLoading) {
            Box(
                modifier = Modifier.fillMaxWidth().padding(32.dp),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator(color = NetflixRed)
            }
        } else if (uiState.sessions.isEmpty()) {
            Text("No active sessions", color = NetflixLightGray)
        } else {
            uiState.sessions.forEach { session ->
                SessionCard(
                    session = session,
                    onRevoke = { viewModel.revokeSession(session.id) },
                )
            }
        }
    }
}

@Composable
private fun SessionCard(session: Session, onRevoke: () -> Unit) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(RoundedCornerShape(4.dp))
                    .background(NetflixGray),
                contentAlignment = Alignment.Center
            ) {
                Icon(LucideIcons.Devices, contentDescription = null, tint = NetflixLightGray, modifier = Modifier.size(20.dp))
            }
            Spacer(modifier = Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(session.deviceName, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                Text(session.ipAddress, color = NetflixLightGray, fontSize = 12.sp)
                Text(
                    if (session.isCurrent) "Current session" else "Last active: ${session.lastActive}",
                    color = NetflixLightGray,
                    fontSize = 12.sp,
                )
            }
            if (!session.isCurrent) {
                Button(
                    onClick = onRevoke,
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                    shape = RoundedCornerShape(4.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 6.dp),
                    modifier = Modifier.height(32.dp)
                ) {
                    Text("Revoke", color = NetflixWhite, fontSize = 14.sp)
                }
            }
        }
    }
}

@Composable
private fun MetadataSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("Metadata", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Configure metadata providers for movies and TV shows", color = NetflixLightGray, fontSize = 14.sp)
        }

        // TMDb Settings
        ProviderSettingsCard(
            name = "The Movie Database (TMDb)",
            description = "Primary metadata provider for movies and TV shows.",
            apiKey = uiState.tmdbSettings?.apiKey ?: "",
            hasBuiltin = uiState.tmdbSettings?.hasBuiltin ?: true,
            isRequired = false,
            placeholder = "Optional",
            url = "https://www.themoviedb.org/settings/api",
            onSave = { apiKey -> viewModel.updateProviderApiKey("tmdb", apiKey) }
        )

        // OMDb Settings
        ProviderSettingsCard(
            name = "Open Movie Database (OMDb)",
            description = "Provides additional ratings and metadata.",
            apiKey = uiState.omdbSettings?.apiKey ?: "",
            hasBuiltin = uiState.omdbSettings?.hasBuiltin ?: false,
            isRequired = true,
            placeholder = "Required",
            url = "https://www.omdbapi.com/apikey.aspx",
            onSave = { apiKey -> viewModel.updateProviderApiKey("omdb", apiKey) }
        )

        // TVDB Settings
        ProviderSettingsCard(
            name = "TheTVDB",
            description = "Fall-back provider for TV shows and episodes.",
            apiKey = uiState.tvdbSettings?.apiKey ?: "",
            hasBuiltin = uiState.tvdbSettings?.hasBuiltin ?: false,
            isRequired = true,
            placeholder = "Required",
            url = "https://thetvdb.com/api-information",
            onSave = { apiKey -> viewModel.updateProviderApiKey("tvdb", apiKey) }
        )

        // Fanart Settings
        ProviderSettingsCard(
            name = "Fanart.tv",
            description = "Provides high quality artwork, clearlogos, and backgrounds.",
            apiKey = uiState.fanartSettings?.apiKey ?: "",
            hasBuiltin = uiState.fanartSettings?.hasBuiltin ?: true,
            isRequired = false,
            placeholder = "Optional",
            url = "https://fanart.tv/get-an-api-key/",
            onSave = { apiKey -> viewModel.updateProviderApiKey("fanart", apiKey) }
        )

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Refresh all metadata", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(4.dp))
                Text("Bulk fetch missing metadata and update ratings for all media.", color = NetflixLightGray, fontSize = 14.sp)
                Spacer(modifier = Modifier.height(16.dp))
                Button(
                    onClick = { viewModel.refreshAllMetadata() },
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Icon(LucideIcons.Refresh, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Refresh Metadata", fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }
}

@Composable
private fun ProviderSettingsCard(
    name: String,
    description: String,
    apiKey: String,
    hasBuiltin: Boolean,
    isRequired: Boolean,
    placeholder: String,
    url: String,
    onSave: (String) -> Unit
) {
    var editedApiKey by androidx.compose.runtime.remember(apiKey) { androidx.compose.runtime.mutableStateOf(apiKey) }
    
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            val statusText = if (apiKey.isNotEmpty()) "Custom Key" else if (hasBuiltin) "Using Internal" else "Not Configured"
            val statusColor = if (apiKey.isNotEmpty()) androidx.compose.ui.graphics.Color(0xFF3B82F6) else if (hasBuiltin) androidx.compose.ui.graphics.Color(0xFF22C55E) else androidx.compose.ui.graphics.Color(0xFF6B7280)
            
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(name, color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                Surface(
                    color = statusColor.copy(alpha = 0.2f),
                    shape = RoundedCornerShape(4.dp),
                ) {
                    Text(
                        text = statusText,
                        color = statusColor,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Medium,
                        modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                    )
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(description, color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(16.dp))
            
            var apiKeyVisible by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

            Text("API / v4 Token", color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = editedApiKey,
                onValueChange = { editedApiKey = it },
                placeholder = { Text(placeholder) },
                modifier = Modifier.fillMaxWidth(),
                visualTransformation = if (apiKeyVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (apiKeyVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { apiKeyVisible = !apiKeyVisible }) { Icon(image, "Toggle API Key visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            Spacer(modifier = Modifier.height(12.dp))
            Button(
                onClick = { onSave(editedApiKey) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            ) {
                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Save", fontWeight = FontWeight.Medium)
            }
        }
    }
}

@OptIn(androidx.compose.foundation.layout.ExperimentalLayoutApi::class)
@Composable
private fun SubtitlesSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("Subtitles", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Configure subtitle providers and auto-download settings", color = NetflixLightGray, fontSize = 14.sp)
        }

        // OpenSubtitles Settings
        OpenSubsSettingsCard(
            settings = uiState.openSubsSettings,
            onSave = { apiKey, username, password ->
                viewModel.updateOpenSubsSettings(
                    com.velox.app.data.model.UpdateOpenSubsRequest(
                        apiKey = apiKey,
                        username = username,
                        password = password.ifEmpty { null }
                    )
                )
            }
        )

        // Subdl Settings
        ProviderSettingsCard(
            name = "Subdl",
            description = "Provides subtitles using the Subdl API.",
            apiKey = uiState.subdlSettings?.apiKey ?: "",
            hasBuiltin = uiState.subdlSettings?.hasBuiltin ?: false,
            isRequired = false,
            placeholder = "Optional",
            url = "https://subdl.com/panel/api",
            onSave = { apiKey -> viewModel.updateProviderApiKey("subdl", apiKey) }
        )

        AITranslationSettingsCard(
            settings = uiState.aiTranslationSettings,
            onSave = { provider, apiKey, baseUrl, model ->
                viewModel.updateAITranslationSettings(
                    provider = provider,
                    apiKey = apiKey,
                    baseUrl = baseUrl,
                    model = model,
                )
            }
        )

        // DeepL Settings
        ProviderSettingsCard(
            name = "DeepL Translation",
            description = "Translate subtitles using DeepL instead of Google Translate.",
            apiKey = uiState.deepLSettings?.apiKey ?: "",
            hasBuiltin = false,
            isRequired = false,
            placeholder = "Leave empty to use Google Translate",
            url = "https://www.deepl.com/pro-api",
            onSave = { apiKey -> viewModel.updateProviderApiKey("deepl", apiKey) }
        )

        // Auto-Download Languages
        AutoSubSettingsCard(
            languages = uiState.autoSubSettings?.languages ?: "",
            onSave = { langs -> viewModel.updateAutoSubSettings(langs) }
        )
    }
}

private fun defaultAITranslationBaseUrl(provider: String): String =
    when (provider) {
        "openai_compatible" -> "https://api.openai.com/v1"
        "gemini_compatible" -> "https://generativelanguage.googleapis.com/v1beta"
        "anthropic_compatible" -> "https://api.anthropic.com"
        else -> ""
    }

@Composable
private fun AITranslationSettingsCard(
    settings: com.velox.app.data.model.AITranslationSettingsDto?,
    onSave: (String, String, String, String) -> Unit,
) {
    var provider by remember(settings?.provider) { mutableStateOf(settings?.provider ?: "") }
    var apiKey by remember(settings?.apiKey) { mutableStateOf(settings?.apiKey ?: "") }
    var baseUrl by remember(settings?.baseUrl, settings?.provider) {
        mutableStateOf(settings?.baseUrl ?: defaultAITranslationBaseUrl(settings?.provider ?: ""))
    }
    var model by remember(settings?.model) { mutableStateOf(settings?.model ?: "") }
    var apiKeyVisible by remember { mutableStateOf(false) }

    val providerOptions = listOf(
        "" to "Disabled",
        "openai_compatible" to "OpenAI Compatible",
        "gemini_compatible" to "Gemini Compatible",
        "anthropic_compatible" to "Anthropic Compatible",
    )
    val isEnabled = provider.isNotEmpty()
    val statusText = providerOptions.find { it.first == provider }?.second ?: "Disabled"
    val statusColor = if (isEnabled) Color(0xFF22C55E) else Color(0xFF6B7280)

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("AI Subtitle Localization", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                Surface(
                    color = statusColor.copy(alpha = 0.2f),
                    shape = RoundedCornerShape(4.dp),
                ) {
                    Text(
                        text = statusText,
                        color = statusColor,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Medium,
                        modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                    )
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                "Use an LLM provider to localize subtitles with better tone and context. Velox manages the system prompt and validates AI JSON output before saving.",
                color = NetflixLightGray,
                fontSize = 12.sp
            )
            Spacer(modifier = Modifier.height(16.dp))

            SettingsDropdown(
                label = "Provider",
                value = provider,
                options = providerOptions,
                onValueChange = { selected ->
                    provider = selected
                    baseUrl = defaultAITranslationBaseUrl(selected)
                },
            )

            Spacer(modifier = Modifier.height(12.dp))
            Text("API Key", color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = apiKey,
                onValueChange = { apiKey = it },
                modifier = Modifier.fillMaxWidth(),
                enabled = isEnabled,
                placeholder = { Text("Provider API key") },
                visualTransformation = if (apiKeyVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (apiKeyVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { apiKeyVisible = !apiKeyVisible }) { Icon(image, "Toggle API Key visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(12.dp))
            OutlinedTextField(
                value = baseUrl,
                onValueChange = { baseUrl = it },
                label = { Text("Base URL") },
                modifier = Modifier.fillMaxWidth(),
                enabled = isEnabled,
                placeholder = { Text(defaultAITranslationBaseUrl(provider)) },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(12.dp))
            OutlinedTextField(
                value = model,
                onValueChange = { model = it },
                label = { Text("Model") },
                modifier = Modifier.fillMaxWidth(),
                enabled = isEnabled,
                placeholder = { Text("e.g. gpt-4.1-mini") },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(8.dp))
            Text(
                "Prompt is system-managed. The response must pass strict JSON validation before a subtitle file is created.",
                color = NetflixLightGray,
                fontSize = 11.sp
            )

            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { onSave(provider, apiKey, baseUrl, model) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            ) {
                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Save", fontWeight = FontWeight.Medium)
            }
        }
    }
}

@Composable
private fun OpenSubsSettingsCard(
    settings: com.velox.app.data.model.OpenSubsSettingsDto?,
    onSave: (String, String, String) -> Unit
) {
    var apiKey by androidx.compose.runtime.remember(settings?.apiKey) { androidx.compose.runtime.mutableStateOf(settings?.apiKey ?: "") }
    var username by androidx.compose.runtime.remember(settings?.username) { androidx.compose.runtime.mutableStateOf(settings?.username ?: "") }
    var password by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf("") }
    var passwordVisible by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            val statusColor = if (settings?.passwordSet == true && apiKey.isNotEmpty()) androidx.compose.ui.graphics.Color(0xFF22C55E) else androidx.compose.ui.graphics.Color(0xFF6B7280)
            
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("OpenSubtitles.com", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                if (settings?.passwordSet == true && apiKey.isNotEmpty()) {
                    Surface(color = statusColor.copy(alpha = 0.2f), shape = RoundedCornerShape(4.dp)) {
                        Text("Enabled", color = statusColor, fontSize = 10.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                    }
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text("Requires both API key and account credentials.", color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(16.dp))

            var apiKeyVisible by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

            // Step 1: API Key
            Row(verticalAlignment = Alignment.CenterVertically) {
                Surface(color = NetflixGray, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(20.dp)) {
                    Box(contentAlignment = Alignment.Center) { Text("1", color = NetflixWhite, fontSize = 10.sp, fontWeight = FontWeight.Bold) }
                }
                Spacer(modifier = Modifier.width(8.dp))
                Text("API Key", color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            }
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = apiKey,
                onValueChange = { apiKey = it },
                modifier = Modifier.fillMaxWidth().padding(start = 28.dp),
                visualTransformation = if (apiKeyVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (apiKeyVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { apiKeyVisible = !apiKeyVisible }) { Icon(image, "Toggle API Key visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            Spacer(modifier = Modifier.height(16.dp))

            // Step 2: Account
            Row(verticalAlignment = Alignment.CenterVertically) {
                Surface(color = NetflixGray, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(20.dp)) {
                    Box(contentAlignment = Alignment.Center) { Text("2", color = NetflixWhite, fontSize = 10.sp, fontWeight = FontWeight.Bold) }
                }
                Spacer(modifier = Modifier.width(8.dp))
                Text("Account", color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            }
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = username,
                onValueChange = { username = it },
                label = { Text("Username") },
                modifier = Modifier.fillMaxWidth().padding(start = 28.dp),
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField(
                value = password,
                onValueChange = { password = it },
                label = { Text(if (settings?.passwordSet == true) "Leave blank to keep current" else "Password") },
                modifier = Modifier.fillMaxWidth().padding(start = 28.dp),
                visualTransformation = if (passwordVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (passwordVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { passwordVisible = !passwordVisible }) { Icon(image, "Toggle password visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { onSave(apiKey, username, password) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
                modifier = Modifier.padding(start = 28.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            ) {
                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Save", fontWeight = FontWeight.Medium)
            }
        }
    }
}

@OptIn(androidx.compose.foundation.layout.ExperimentalLayoutApi::class)
@Composable
private fun AutoSubSettingsCard(
    languages: String,
    onSave: (String) -> Unit
) {
    val commonLangs = listOf("en" to "English", "vi" to "Vietnamese", "fr" to "French", "de" to "German", "es" to "Spanish", "ja" to "Japanese", "ko" to "Korean", "zh" to "Chinese")
    val selectedLangs = androidx.compose.runtime.remember(languages) { androidx.compose.runtime.mutableStateListOf(*languages.split(",").filter { it.isNotEmpty() }.toTypedArray()) }

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("Auto-Download Languages", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                if (selectedLangs.isNotEmpty()) {
                    Surface(color = androidx.compose.ui.graphics.Color(0xFF3B82F6).copy(alpha = 0.2f), shape = RoundedCornerShape(4.dp)) {
                        Text("${selectedLangs.size} languages", color = androidx.compose.ui.graphics.Color(0xFF3B82F6), fontSize = 10.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                    }
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text("Select target languages for automatic subtitle downloads.", color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(12.dp))

            androidx.compose.foundation.layout.FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                for (lang in commonLangs) {
                    val code = lang.first
                    val label = lang.second
                    val isSelected = selectedLangs.contains(code)
                    
                    Surface(
                        color = if (isSelected) NetflixRed else NetflixGray,
                        shape = androidx.compose.foundation.shape.CircleShape,
                        modifier = Modifier.clickable {
                            if (isSelected) selectedLangs.remove(code) else selectedLangs.add(code)
                        }
                    ) {
                        Text(
                            label,
                            color = if (isSelected) NetflixWhite else NetflixLightGray,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Medium,
                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp)
                        )
                    }
                }
            }
            
            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { onSave(selectedLangs.joinToString(",")) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            ) {
                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Save", fontWeight = FontWeight.Medium)
            }
        }
    }
}

@Composable
private fun PlaybackSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("Playback", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Configure global playback streaming policy", color = NetflixLightGray, fontSize = 14.sp)
        }

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Playback Mode", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(4.dp))
                Text("Controls whether the server should automatically optimize playback or force direct streaming when possible.", color = NetflixLightGray, fontSize = 12.sp)
                Spacer(modifier = Modifier.height(16.dp))

                val currentMode = uiState.adminPlaybackSettings?.playbackMode ?: "auto"

                // Auto Option
                Surface(
                    modifier = Modifier.fillMaxWidth().clickable { viewModel.updatePlaybackMode("auto") },
                    color = if (currentMode == "auto") androidx.compose.ui.graphics.Color(0xFFE50914).copy(alpha = 0.1f) else androidx.compose.ui.graphics.Color.Transparent,
                    shape = RoundedCornerShape(8.dp),
                    border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "auto") NetflixRed else NetflixGray.copy(alpha = 0.5f))
                ) {
                    Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.Top) {
                        Surface(
                            shape = androidx.compose.foundation.shape.CircleShape,
                            color = if (currentMode == "auto") NetflixRed else androidx.compose.ui.graphics.Color.Transparent,
                            border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "auto") NetflixRed else androidx.compose.ui.graphics.Color.Gray),
                            modifier = Modifier.size(20.dp).padding(top = 2.dp)
                        ) {
                            if (currentMode == "auto") {
                                Box(contentAlignment = Alignment.Center) {
                                    Surface(color = NetflixWhite, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(8.dp)) {}
                                }
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text("Automatic", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            Spacer(modifier = Modifier.height(4.dp))
                            Text("Automatically adapt quality based on network conditions and device capabilities. Best for general use.", color = NetflixLightGray, fontSize = 12.sp)
                        }
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                // Direct Play Option
                Surface(
                    modifier = Modifier.fillMaxWidth().clickable { viewModel.updatePlaybackMode("direct_play") },
                    color = if (currentMode == "direct_play") androidx.compose.ui.graphics.Color(0xFFE50914).copy(alpha = 0.1f) else androidx.compose.ui.graphics.Color.Transparent,
                    shape = RoundedCornerShape(8.dp),
                    border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "direct_play") NetflixRed else NetflixGray.copy(alpha = 0.5f))
                ) {
                    Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.Top) {
                        Surface(
                            shape = androidx.compose.foundation.shape.CircleShape,
                            color = if (currentMode == "direct_play") NetflixRed else androidx.compose.ui.graphics.Color.Transparent,
                            border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "direct_play") NetflixRed else androidx.compose.ui.graphics.Color.Gray),
                            modifier = Modifier.size(20.dp).padding(top = 2.dp)
                        ) {
                            if (currentMode == "direct_play") {
                                Box(contentAlignment = Alignment.Center) {
                                    Surface(color = NetflixWhite, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(8.dp)) {}
                                }
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text("Force Direct Play", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            Spacer(modifier = Modifier.height(4.dp))
                            Text("Disable transcoding gracefully. Media will be streamed directly when possible.", color = NetflixLightGray, fontSize = 12.sp)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun CinemaSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("Cinema Mode", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Configure trailer playback and custom intros before your movies and TV shows.", color = NetflixLightGray, fontSize = 14.sp)
        }

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = androidx.compose.ui.graphics.Color(0xFFEAB308).copy(alpha = 0.1f),
            shape = RoundedCornerShape(8.dp),
            border = androidx.compose.foundation.BorderStroke(1.dp, androidx.compose.ui.graphics.Color(0xFFEAB308).copy(alpha = 0.5f))
        ) {
            Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                Icon(LucideIcons.Info, contentDescription = null, tint = androidx.compose.ui.graphics.Color(0xFFEAB308), modifier = Modifier.size(24.dp))
                Spacer(modifier = Modifier.width(12.dp))
                Text("Please note: Cinema Mode is currently only supported on the Web application and via Casting. These settings will not affect playback on this Android device.", color = androidx.compose.ui.graphics.Color(0xFFFEF08A), fontSize = 13.sp)
            }
        }

        // Enable Cinema Mode
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Enable Cinema Mode", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text("Play trailers from upcoming movies or your own library before the main feature.", color = NetflixLightGray, fontSize = 12.sp)
                    }
                    Switch(
                        checked = uiState.adminCinemaSettings?.enabled ?: false,
                        onCheckedChange = { viewModel.updateCinemaSettings(enabled = it) },
                        colors = SwitchDefaults.colors(
                            checkedThumbColor = NetflixWhite,
                            checkedTrackColor = NetflixRed,
                            uncheckedThumbColor = NetflixLightGray,
                            uncheckedTrackColor = NetflixGray
                        )
                    )
                }
            }
        }

        // Max Trailers
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Max Trailers", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(4.dp))
                Text("Maximum number of trailers to play before the main content.", color = NetflixLightGray, fontSize = 12.sp)
                Spacer(modifier = Modifier.height(16.dp))

                SettingsDropdown(
                    label = "",
                    value = uiState.adminCinemaSettings?.maxTrailers ?: "2",
                    options = listOf("0" to "Original", "1" to "1 Trailer", "2" to "2 Trailers", "3" to "3 Trailers"),
                    onValueChange = { viewModel.updateCinemaSettings(maxTrailers = it) }
                )
            }
        }

        // Custom Intro
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                    Text("Custom Intro", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    if (uiState.adminCinemaSettings?.hasIntro == true) {
                        Surface(color = androidx.compose.ui.graphics.Color(0xFF22C55E).copy(alpha = 0.2f), shape = RoundedCornerShape(4.dp)) {
                            Text("Intro uploaded", color = androidx.compose.ui.graphics.Color(0xFF22C55E), fontSize = 10.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                        }
                    }
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text("Upload a custom intro video to play before movies.", color = NetflixLightGray, fontSize = 12.sp)
                Spacer(modifier = Modifier.height(16.dp))
                
                Button(
                    onClick = { /* File picker integration required */ },
                    colors = ButtonDefaults.buttonColors(containerColor = androidx.compose.ui.graphics.Color(0xFF2563EB)),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Icon(LucideIcons.Upload, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(if (uiState.adminCinemaSettings?.hasIntro == true) "Upload New Intro" else "Upload Intro", fontWeight = FontWeight.Medium)
                }
            }
        }
    }
}

private fun formatBytes(bytes: Long): String {
    if (bytes <= 0) return "0 B"
    val units = arrayOf("B", "KB", "MB", "GB", "TB")
    val digitGroups = (Math.log10(bytes.toDouble()) / Math.log10(1024.0)).toInt()
    return String.format("%.1f %s", bytes / Math.pow(1024.0, digitGroups.toDouble()), units[digitGroups])
}

@Composable
private fun PretranscodeSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    val settings = uiState.pretranscodeSettings
    val status = uiState.pretranscodeStatus
    val profiles = uiState.pretranscodeProfiles
    val estimate = uiState.pretranscodeEstimate
    val libraries = uiState.libraries

    var selectedLibraryId by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(0) }
    var showCleanupConfirm by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

    // Auto-select first library
    androidx.compose.runtime.LaunchedEffect(libraries, selectedLibraryId) {
        if (selectedLibraryId == 0 && libraries.isNotEmpty()) {
            selectedLibraryId = libraries.first().id
            viewModel.loadPretranscodeEstimate(selectedLibraryId)
        }
    }

    val isEncoding = status != null && (status.encoding > 0 || status.queued > 0) && !status.paused
    val progressPercent = if (status != null && status.total > 0) ((status.done.toFloat() / status.total.toFloat()) * 100).toInt() else 0

    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("Pre-transcode", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Encode media in advance for instant playback — no buffering, no waiting.", color = NetflixLightGray, fontSize = 14.sp)
        }

        // Enable Toggle
        Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Offline Encoding", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text("Pre-encode your library into browser-compatible H.264+AAC MP4 files. Like Netflix — instant playback, zero transcoding delay.", color = NetflixLightGray, fontSize = 12.sp)
                    }
                    Switch(
                        checked = settings?.enabled == true,
                        onCheckedChange = { viewModel.updatePretranscodeSettings(enabled = it) },
                        colors = SwitchDefaults.colors(checkedThumbColor = NetflixWhite, checkedTrackColor = NetflixRed)
                    )
                }
            }
        }

        if (settings?.enabled == true) {
            // Quality Profiles
            Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Quality Profiles", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    profiles.forEach { p ->
                        Surface(
                            color = if (p.enabled) NetflixRed.copy(alpha = 0.1f) else NetflixGray.copy(alpha = 0.2f),
                            shape = RoundedCornerShape(8.dp),
                            modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp).clickable { viewModel.togglePretranscodeProfile(p.id, !p.enabled) }
                        ) {
                            Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                                Checkbox(
                                    checked = p.enabled,
                                    onCheckedChange = { viewModel.togglePretranscodeProfile(p.id, it) },
                                    colors = androidx.compose.material3.CheckboxDefaults.colors(checkedColor = NetflixRed)
                                )
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(p.name, color = NetflixWhite, fontWeight = FontWeight.Medium, fontSize = 14.sp)
                                    Text("${p.height}p, ${p.videoBitrate / 1000}Mbps video + ${p.audioBitrate}kbps audio", color = NetflixLightGray, fontSize = 12.sp)
                                }
                                val gbPerFilm = (((p.videoBitrate + p.audioBitrate) * 5400) / 8 / 1024 / 1024.0).toFloat()
                                Text(String.format("~%.1f GB/film", gbPerFilm), color = NetflixLightGray, fontSize = 12.sp)
                            }
                        }
                    }
                }
            }

            // Schedule & Concurrency
            Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Schedule", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    SettingsDropdown(
                        label = "Encode time",
                        value = settings.schedule.takeIf { it.isNotEmpty() } ?: "always",
                        options = listOf("always" to "Always (fastest)", "night" to "Night only (00:00–06:00)", "idle" to "When idle (no one watching)"),
                        onValueChange = { viewModel.updatePretranscodeSettings(schedule = it) }
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    SettingsDropdown(
                        label = "Concurrent jobs",
                        value = settings.concurrency.takeIf { it.isNotEmpty() } ?: "1",
                        options = listOf("1" to "1 (NAS-friendly)", "2" to "2", "3" to "3", "4" to "4 (powerful server)"),
                        onValueChange = { viewModel.updatePretranscodeSettings(concurrency = it) }
                    )
                }
            }

            // Storage Estimation
            Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Storage Estimation", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    val libraryOptions = libraries.map { it.id.toString() to it.name }
                    SettingsDropdown(
                        label = "Library",
                        value = selectedLibraryId.toString(),
                        options = libraryOptions.ifEmpty { listOf("0" to "None") },
                        onValueChange = { 
                            val newId = it.toIntOrNull() ?: 0
                            selectedLibraryId = newId
                            viewModel.loadPretranscodeEstimate(newId)
                        }
                    )
                    
                    if (estimate != null) {
                        Spacer(modifier = Modifier.height(16.dp))
                        Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            estimate.profiles?.forEach { ep ->
                                Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                                    Text("${ep.profileName} (${ep.fileCount} files)", color = NetflixLightGray, fontSize = 14.sp)
                                    Text(String.format("%.1f GB", ep.estimatedGb), color = NetflixWhite, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                                }
                            }
                            HorizontalDivider(color = NetflixGray, modifier = Modifier.padding(vertical = 4.dp))
                            Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                                Text("Total", color = NetflixWhite, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                                Text(formatBytes(estimate.totalBytes), color = NetflixWhite, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                            }
                            Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                                Text("Disk free", color = NetflixLightGray, fontSize = 14.sp)
                                val isEnough = estimate.diskFreeBytes > estimate.totalBytes
                                val color = if (isEnough) androidx.compose.ui.graphics.Color(0xFF22C55E) else androidx.compose.ui.graphics.Color(0xFFEF4444)
                                Text("${formatBytes(estimate.diskFreeBytes)} " + if (isEnough) "✓ Enough" else "✗ Not enough", color = color, fontSize = 14.sp)
                            }
                        }
                    }
                }
            }

            // Progress Dashboard
            if (status != null && status.total > 0) {
                Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text("Progress", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                        Spacer(modifier = Modifier.height(16.dp))
                        
                        Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                            Text("${status.done}/${status.total} files", color = NetflixLightGray, fontSize = 12.sp)
                            Text("$progressPercent%", color = NetflixLightGray, fontSize = 12.sp)
                        }
                        Spacer(modifier = Modifier.height(4.dp))
                        LinearProgressIndicator(
                            progress = { progressPercent.toFloat() / 100f },
                            modifier = Modifier.fillMaxWidth().height(8.dp).clip(RoundedCornerShape(4.dp)),
                            color = NetflixRed,
                            trackColor = NetflixGray,
                            drawStopIndicator = {}
                        )
                        
                        if (status.currentFile != null) {
                            Spacer(modifier = Modifier.height(12.dp))
                            Row {
                                Text("Encoding: ", color = NetflixLightGray, fontSize = 12.sp)
                                Text(status.currentFile, color = NetflixWhite, fontSize = 12.sp, maxLines = 1, modifier = Modifier.weight(1f))
                                if (status.speed != null) {
                                    Text("(${status.speed})", color = NetflixLightGray, fontSize = 12.sp)
                                }
                            }
                        }
                        
                        Spacer(modifier = Modifier.height(16.dp))
                        Row(horizontalArrangement = Arrangement.SpaceEvenly, modifier = Modifier.fillMaxWidth()) {
                            Text("✓ Done: ${status.done}", color = androidx.compose.ui.graphics.Color(0xFF22C55E), fontSize = 12.sp)
                            if (status.failed > 0) Text("✗ Failed: ${status.failed}", color = androidx.compose.ui.graphics.Color(0xFFEF4444), fontSize = 12.sp)
                            Text("Queued: ${status.queued}", color = NetflixLightGray, fontSize = 12.sp)
                            Text("Disk: ${formatBytes(status.diskUsed)}", color = NetflixLightGray, fontSize = 12.sp)
                        }
                        
                        Spacer(modifier = Modifier.height(16.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                            if (status.paused) {
                                Button(
                                    onClick = { viewModel.executePretranscodeAction("resume") },
                                    colors = ButtonDefaults.buttonColors(containerColor = androidx.compose.ui.graphics.Color(0xFF22C55E)),
                                    shape = RoundedCornerShape(8.dp)
                                ) {
                                    Icon(LucideIcons.PlayArrow, contentDescription = null, modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("Resume")
                                }
                            } else if (isEncoding) {
                                Button(
                                    onClick = { viewModel.executePretranscodeAction("stop") /* Note: Stop acts as pause/stop currently in Velox server */ },
                                    colors = ButtonDefaults.buttonColors(containerColor = androidx.compose.ui.graphics.Color(0xFFEAB308)),
                                    shape = RoundedCornerShape(8.dp)
                                ) {
                                    Icon(LucideIcons.Pause, contentDescription = null, modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("Pause")
                                }
                            }
                            
                            Button(
                                onClick = { viewModel.executePretranscodeAction("stop") },
                                colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                                shape = RoundedCornerShape(8.dp)
                            ) {
                                Icon(androidx.compose.material.icons.Icons.Filled.Close, contentDescription = null, modifier = Modifier.size(16.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Stop")
                            }
                        }
                    }
                }
            }

            // Actions
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                if (!isEncoding) {
                    Button(
                        onClick = { viewModel.executePretranscodeAction("start") },
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text("Start Encoding", fontWeight = FontWeight.SemiBold)
                    }
                }
                
                Button(
                    onClick = { showCleanupConfirm = true },
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Icon(LucideIcons.Delete, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Delete All Files")
                }
            }
        }
    }

    if (showCleanupConfirm) {
        AlertDialog(
            onDismissRequest = { showCleanupConfirm = false },
            title = { Text("Delete All Pre-transcode Files", color = NetflixWhite) },
            text = { Text("This will permanently delete all pre-encoded files and disable pre-transcode. Your original media files are NOT affected.", color = NetflixLightGray) },
            containerColor = NetflixDark,
            confirmButton = {
                TextButton(onClick = { 
                    viewModel.executePretranscodeAction("cleanup")
                    showCleanupConfirm = false 
                }) {
                    Text("Delete All", color = androidx.compose.ui.graphics.Color(0xFFEF4444))
                }
            },
            dismissButton = {
                TextButton(onClick = { showCleanupConfirm = false }) {
                    Text("Cancel", color = NetflixWhite)
                }
            }
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun MarkersSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    var selectedLibraryId by remember { mutableStateOf(0) }

    // Auto-select first library (or fix stale selection)
    LaunchedEffect(uiState.libraries) {
        if (uiState.libraries.isNotEmpty() && (selectedLibraryId == 0 || uiState.libraries.none { it.id == selectedLibraryId })) {
            selectedLibraryId = uiState.libraries.first().id
        }
    }

    val libraryOptions = uiState.libraries.map { it.id.toString() to it.name }

    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        // Section header (matches web SectionHeader)
        Column {
            Text("Skip Intro / Credits", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Detect and manage intro/credits skip markers", color = NetflixLightGray, fontSize = 14.sp)
        }

        // ── Stats Overview Card ──
        val stats = uiState.markerStats
        MarkerCard {
            Text("Overview", color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(16.dp))

            if (uiState.isLoading) {
                Text("Loading...", color = NetflixLightGray, fontSize = 14.sp)
            } else if (stats != null && stats.totalMarkers > 0) {
                // 4-column stat grid
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    MarkerStatBlock(
                        value = stats.totalMarkers.toString(),
                        label = "Total Markers",
                        valueColor = NetflixWhite,
                        modifier = Modifier.weight(1f)
                    )
                    MarkerStatBlock(
                        value = stats.introMarkers.toString(),
                        label = "Intro Markers",
                        valueColor = Color(0xFF60A5FA), // blue-400
                        modifier = Modifier.weight(1f)
                    )
                    MarkerStatBlock(
                        value = stats.creditsMarkers.toString(),
                        label = "Credits Markers",
                        valueColor = Color(0xFFC084FC), // purple-400
                        modifier = Modifier.weight(1f)
                    )
                    MarkerStatBlock(
                        value = stats.totalFiles.toString(),
                        label = "Total Files",
                        valueColor = Color(0xFFD1D5DB), // gray-300
                        modifier = Modifier.weight(1f)
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Coverage bars
                val introPercent = if (stats.totalFiles > 0) ((stats.filesWithIntro.toFloat() / stats.totalFiles) * 100).toInt() else 0
                val creditsPercent = if (stats.totalFiles > 0) ((stats.filesWithCredits.toFloat() / stats.totalFiles) * 100).toInt() else 0

                MarkerCoverageBar(
                    label = "Files with Intro (${stats.filesWithIntro}/${stats.totalFiles})",
                    percent = introPercent,
                    color = Color(0xFF3B82F6) // blue-500
                )

                Spacer(modifier = Modifier.height(12.dp))

                MarkerCoverageBar(
                    label = "Files with Credits (${stats.filesWithCredits}/${stats.totalFiles})",
                    percent = creditsPercent,
                    color = Color(0xFFA855F7) // purple-500
                )

                // Source breakdown
                Spacer(modifier = Modifier.height(20.dp))
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                Spacer(modifier = Modifier.height(16.dp))

                Text("Source Breakdown", color = Color(0xFFD1D5DB), fontSize = 14.sp, fontWeight = FontWeight.Medium)
                Spacer(modifier = Modifier.height(8.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
                    if (stats.chapterSource > 0) {
                        Text("Chapter: ${stats.chapterSource}", color = Color(0xFF4ADE80), fontSize = 14.sp)
                    }
                    if (stats.fingerprintSource > 0) {
                        Text("Fingerprint: ${stats.fingerprintSource}", color = Color(0xFF60A5FA), fontSize = 14.sp)
                    }
                    if (stats.manualSource > 0) {
                        Text("Manual: ${stats.manualSource}", color = Color(0xFFFACC15), fontSize = 14.sp)
                    }
                }
            } else {
                // Empty state — matches web's "No markers detected yet"
                Text(
                    "No markers detected yet. Run detection below to scan your library.",
                    color = Color(0xFF6B7280), // gray-500
                    fontSize = 14.sp
                )
            }
        }

        // ── Run Detection Card ──
        MarkerCard {
            Text("Detection", color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                "Run audio fingerprint analysis to detect intro/credits segments across your library. Works best for TV series with recurring intros.",
                color = NetflixLightGray,
                fontSize = 14.sp
            )

            Spacer(modifier = Modifier.height(20.dp))

            // Reuse SettingsDropdown component (ModalBottomSheet picker)
            SettingsDropdown(
                label = "Select Library",
                value = selectedLibraryId.toString(),
                options = libraryOptions,
                onValueChange = { selectedLibraryId = it.toIntOrNull() ?: 0 },
            )

            Spacer(modifier = Modifier.height(16.dp))

            Button(
                onClick = { viewModel.runMarkerDetection(selectedLibraryId) },
                colors = ButtonDefaults.buttonColors(
                    containerColor = NetflixRed,
                    contentColor = NetflixWhite,
                    disabledContainerColor = NetflixRed.copy(alpha = 0.5f)
                ),
                shape = RoundedCornerShape(8.dp),
                enabled = !uiState.isRunningDetection && selectedLibraryId != 0,
                modifier = Modifier.fillMaxWidth().height(48.dp)
            ) {
                if (uiState.isRunningDetection) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(18.dp),
                        color = NetflixWhite,
                        strokeWidth = 2.dp
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Running...", fontWeight = FontWeight.SemiBold)
                } else {
                    Text("Run Detection", fontWeight = FontWeight.SemiBold)
                }
            }

            // Running indicator — matches web's real-time progress area
            if (uiState.isRunningDetection) {
                Spacer(modifier = Modifier.height(16.dp))
                Surface(
                    modifier = Modifier.fillMaxWidth(),
                    color = NetflixGray,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(16.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            color = NetflixRed,
                            strokeWidth = 2.dp
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text("Analyzing library...", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            Text("Audio fingerprint detection in progress", color = NetflixLightGray, fontSize = 12.sp)
                        }
                    }
                }
            }

            // Success result — matches web's complete result panel
            uiState.successMessage?.let { message ->
                if (!uiState.isRunningDetection) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = NetflixGray,
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.CheckCircle,
                                contentDescription = null,
                                tint = Color(0xFF4ADE80),
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(12.dp))
                            Text(message, color = Color(0xFF4ADE80), fontSize = 14.sp)
                        }
                    }
                }
            }

            // Error result — matches web's error panel
            uiState.error?.let { errorMsg ->
                if (!uiState.isRunningDetection) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = Color(0xFF7F1D1D).copy(alpha = 0.3f), // red-900/30
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.Warning,
                                contentDescription = null,
                                tint = Color(0xFFF87171), // red-400
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(12.dp))
                            Text(errorMsg, color = Color(0xFFF87171), fontSize = 14.sp)
                        }
                    }
                }
            }
        }

        // ── How it works Card ──
        MarkerCard {
            Text("How it works", color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(16.dp))

            MarkerBulletItem(Color(0xFF4ADE80), "Chapter — Extracted from video chapter metadata during scan")
            Spacer(modifier = Modifier.height(12.dp))
            MarkerBulletItem(Color(0xFF60A5FA), "Fingerprint — Audio analysis compares episodes to find recurring segments")
            Spacer(modifier = Modifier.height(12.dp))
            MarkerBulletItem(Color(0xFFFACC15), "Manual — User-adjusted skip boundaries manually set in player")
        }
    }
}

/** Card wrapper matching web's `rounded-lg bg-netflix-black p-6 ring-1 ring-white/10` */
@Composable
private fun MarkerCard(content: @Composable ColumnScope.() -> Unit) {
    Surface(
        modifier = Modifier
            .fillMaxWidth(),
        color = NetflixBlack,
        shape = RoundedCornerShape(12.dp),
        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.1f))
    ) {
        Column(modifier = Modifier.padding(20.dp)) {
            content()
        }
    }
}

@Composable
private fun MarkerStatBlock(value: String, label: String, valueColor: Color, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .background(NetflixGray, RoundedCornerShape(8.dp))
            .padding(12.dp)
    ) {
        Text(value, color = valueColor, fontSize = 24.sp, fontWeight = FontWeight.Bold)
        Spacer(modifier = Modifier.height(4.dp))
        Text(label, color = NetflixLightGray, fontSize = 11.sp, lineHeight = 14.sp)
    }
}

@Composable
private fun MarkerCoverageBar(label: String, percent: Int, color: Color) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(label, color = NetflixLightGray, fontSize = 14.sp)
            Text("$percent%", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
        }
        Spacer(modifier = Modifier.height(4.dp))
        LinearProgressIndicator(
            progress = { percent / 100f },
            modifier = Modifier
                .fillMaxWidth()
                .height(8.dp)
                .clip(RoundedCornerShape(4.dp)),
            color = color,
            trackColor = Color(0xFF374151), // gray-700
            drawStopIndicator = {}
        )
    }
}

@Composable
private fun MarkerBulletItem(color: Color, text: String) {
    Row(verticalAlignment = Alignment.Top) {
        Box(
            modifier = Modifier
                .padding(top = 6.dp, end = 8.dp)
                .size(8.dp)
                .background(color, androidx.compose.foundation.shape.CircleShape)
        )
        Text(text, color = NetflixLightGray, fontSize = 14.sp, lineHeight = 20.sp)
    }
}

@Composable
private fun DashboardSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    val serverInfo = uiState.serverInfo
    val libraryStats = uiState.libraryStats

    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        // Section header (matches web SectionHeader)
        Column {
            Text("Dashboard", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Server information and status", color = NetflixLightGray, fontSize = 14.sp)
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(color = NetflixRed, modifier = Modifier.padding(16.dp))
            return
        }

        // Stats Cards
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            DashboardStatCard("Total Media", serverInfo?.mediaCount?.toString() ?: "0", modifier = Modifier.weight(1f))
            DashboardStatCard("Series", serverInfo?.seriesCount?.toString() ?: "0", modifier = Modifier.weight(1f))
            DashboardStatCard("Users", serverInfo?.userCount?.toString() ?: "0", modifier = Modifier.weight(1f))
            DashboardStatCard("Total Size", formatBytes(serverInfo?.totalSize ?: 0L), modifier = Modifier.weight(1f))
        }

        // Server Information
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(8.dp),
        ) {
            Column(modifier = Modifier.padding(20.dp)) {
                Text("Server Information", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(16.dp))
                
                DashboardInfoRow("Version", serverInfo?.version ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("Uptime", serverInfo?.uptime ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("Go Version", serverInfo?.goVersion ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("OS / Arch", "${serverInfo?.os ?: "?"} / ${serverInfo?.arch ?: "?"}")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("FFmpeg", serverInfo?.ffmpegVersion ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("HW Acceleration", serverInfo?.hwAccel?.takeIf { it.isNotBlank() } ?: "None")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("Database", serverInfo?.database ?: "SQLite")
            }
        }

        // Library Statistics
        if (libraryStats.isNotEmpty()) {
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column {
                    Text(
                        "Library Statistics",
                        color = NetflixWhite,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(20.dp)
                    )
                    
                    // Header Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color.Black.copy(alpha = 0.5f))
                            .padding(horizontal = 20.dp, vertical = 10.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Library", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(1.5f))
                        Text("Type", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(1f))
                        Text("Items", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(0.5f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                        Text("Size", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(1.2f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                    }
                    
                    // List Rows
                    libraryStats.forEachIndexed { index, lib ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 20.dp, vertical = 12.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(lib.name, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                            
                            val typeColor = when(lib.type.lowercase(java.util.Locale.US)) {
                                "movies" -> Color(0xFF60A5FA)
                                "tvshows" -> Color(0xFFC084FC)
                                else -> Color(0xFF4ADE80)
                            }
                            Text(lib.type, color = typeColor, fontSize = 12.sp, modifier = Modifier.weight(1f))
                            
                            Text(lib.itemCount.toString(), color = Color(0xFFD1D5DB), fontSize = 14.sp, modifier = Modifier.weight(0.5f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                            Text(formatBytes(lib.totalSize), color = Color(0xFFD1D5DB), fontSize = 14.sp, modifier = Modifier.weight(1.2f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                        }
                        if (index < libraryStats.lastIndex) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun DashboardStatCard(label: String, value: String, modifier: Modifier = Modifier) {
    Surface(
        modifier = modifier,
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.Start,
        ) {
            Text(label, color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(4.dp))
            Text(value, color = NetflixWhite, fontSize = 24.sp, fontWeight = FontWeight.Bold, maxLines = 1)
        }
    }
}

@Composable
private fun DashboardInfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(label, color = NetflixLightGray, fontSize = 14.sp)
        Text(value, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
    }
}

@Composable
private fun LibrariesSection(
    viewModel: SettingsViewModel?,
    uiState: SettingsUiState,
    onAddClick: () -> Unit,
    onEditClick: (com.velox.app.domain.model.Library) -> Unit,
    onDeleteClick: (Int) -> Unit,
    onScanClick: (Int, Boolean) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column {
                Text("Libraries", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                val count = uiState.libraries.size
                Text("$count ${if (count == 1) "library" else "libraries"} configured", color = NetflixLightGray, fontSize = 14.sp)
            }
            Button(
                onClick = onAddClick,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Add, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Add Library", fontWeight = FontWeight.SemiBold)
            }
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(color = NetflixRed, modifier = Modifier.padding(16.dp))
        } else if (uiState.libraries.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Icon(LucideIcons.Movie, contentDescription = null, tint = NetflixGray, modifier = Modifier.size(36.dp))
                    Spacer(modifier = Modifier.height(8.dp))
                    Text("No libraries configured", color = NetflixLightGray, fontSize = 14.sp)
                    Spacer(modifier = Modifier.height(12.dp))
                    Button(
                        onClick = onAddClick,
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        shape = RoundedCornerShape(8.dp),
                    ) {
                        Text("Add Library", fontWeight = FontWeight.Medium)
                    }
                }
            }
        } else {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                uiState.libraries.forEach { library ->
                    LibraryCard(
                        library = library,
                        isScanning = uiState.scanningLibraryId == library.id,
                        onScan = { onScanClick(library.id, false) },
                        onForceScan = { onScanClick(library.id, true) },
                        onDelete = { onDeleteClick(library.id) }
                    )
                }
            }
        }
    }
}

@Composable
private fun LibraryCard(
    library: com.velox.app.domain.model.Library,
    isScanning: Boolean,
    onScan: () -> Unit,
    onForceScan: () -> Unit,
    onDelete: () -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            // Left Content
            Row(
                modifier = Modifier.weight(1f),
                verticalAlignment = Alignment.CenterVertically
            ) {
                val typeDetails = when(library.type.lowercase(java.util.Locale.US)) {
                    "movies" -> Triple(LucideIcons.Movie, Color(0xFF3B82F6), Color(0xFF1E3A8A)) // Blue
                    "tvshows" -> Triple(LucideIcons.Tv, Color(0xFFA855F7), Color(0xFF4C1D95)) // Purple
                    else -> Triple(LucideIcons.ListIcon, Color(0xFF22C55E), Color(0xFF14532D)) // Green
                }
                
                // Icon Box
                Surface(
                    modifier = Modifier.size(40.dp),
                    shape = RoundedCornerShape(8.dp),
                    color = typeDetails.third.copy(alpha = 0.4f)
                ) {
                    Box(contentAlignment = Alignment.Center) {
                        Icon(typeDetails.first, contentDescription = null, tint = typeDetails.second, modifier = Modifier.size(20.dp))
                    }
                }
                
                Spacer(modifier = Modifier.width(12.dp))
                
                Column {
                    Text(library.name, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(2.dp))
                    library.paths.takeTop(3).forEach { p ->
                        Text(p, color = NetflixLightGray, fontSize = 12.sp, fontFamily = androidx.compose.ui.text.font.FontFamily.Monospace, maxLines = 1)
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    val prettyType = when(library.type.lowercase(java.util.Locale.US)) {
                        "movies" -> "Movies"
                        "tvshows" -> "TV Shows"
                        else -> "Mixed"
                    }
                    Surface(
                        shape = RoundedCornerShape(4.dp),
                        color = Color.Transparent,
                        border = androidx.compose.foundation.BorderStroke(1.dp, typeDetails.second)
                    ) {
                        Text(prettyType, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp), color = typeDetails.second, fontSize = 10.sp)
                    }
                }
            }
            
            // Right Buttons
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                LibraryCardAction(
                    text = "Scan",
                    icon = LucideIcons.Refresh,
                    onClick = onScan,
                    isLoading = isScanning,
                    disabled = isScanning,
                )
                
                LibraryCardAction(
                    text = "Force Rescan",
                    icon = LucideIcons.Refresh,
                    onClick = onForceScan,
                    disabled = isScanning,
                )
                
                LibraryCardAction(
                    text = null,
                    icon = LucideIcons.Delete,
                    onClick = onDelete,
                    disabled = isScanning,
                )
            }
        }
    }
}

@Composable
private fun LibraryCardAction(
    text: String?,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
    isLoading: Boolean = false,
    disabled: Boolean = false,
) {
    Surface(
        shape = RoundedCornerShape(4.dp),
        color = NetflixGray,
        modifier = Modifier
            .height(32.dp)
            .clip(RoundedCornerShape(4.dp))
            .clickable(onClick = onClick, enabled = !disabled)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = if (text != null) 12.dp else 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center
        ) {
            if (isLoading) {
                CircularProgressIndicator(
                    color = NetflixWhite,
                    strokeWidth = 2.dp,
                    modifier = Modifier.size(13.dp)
                )
            } else {
                Icon(icon, contentDescription = null, modifier = Modifier.size(14.dp), tint = if (disabled) NetflixLightGray else NetflixWhite)
            }
            
            if (text != null) {
                Spacer(modifier = Modifier.width(6.dp))
                val label = if (isLoading && text == "Scan") "Scanning" else text
                Text(
                    text = label, 
                    fontSize = 12.sp, 
                    color = if (disabled && !isLoading) NetflixLightGray else NetflixWhite, 
                    fontWeight = FontWeight.Medium
                )
            }
        }
    }
}

private fun <T> List<T>.takeTop(n: Int): List<T> = if (size > n) take(n) else this

@Composable
private fun UsersSection(
    viewModel: SettingsViewModel?,
    uiState: SettingsUiState,
    onAddClick: () -> Unit,
    onEditClick: (AdminUser) -> Unit,
    onDeleteClick: (Int) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column {
                Text("Users", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                val count = uiState.users.size
                Text("$count ${if (count == 1) "user" else "users"}", color = NetflixLightGray, fontSize = 14.sp)
            }
            Button(
                onClick = onAddClick,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Add, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Add User", fontWeight = FontWeight.SemiBold)
            }
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(color = NetflixRed, modifier = Modifier.padding(16.dp))
        } else if (uiState.users.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text("No users found", color = NetflixLightGray, fontSize = 14.sp)
                }
            }
        } else {
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column {
                    // Header Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("User", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2.5f))
                        Text("Role", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.2f))
                        Text("Created", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                        Text("Actions", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
                    }
                    
                    HorizontalDivider(color = Color.White.copy(alpha = 0.05f))

                    // Users List
                    uiState.users.forEachIndexed { index, user ->
                        if (index > 0) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.02f))
                        }
                        UserTableRow(
                            user = user,
                            canDelete = uiState.user?.id != user.id,
                            onEdit = { onEditClick(user) },
                            onDelete = { onDeleteClick(user.id) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun UserTableRow(
    user: AdminUser,
    canDelete: Boolean,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
) {
    val displayName = user.displayName.ifBlank { user.username }
    val initial = displayName.take(1).uppercase()
    
    val rawDate = user.createdAt.substringBefore("T") 
    val fmtDate = try {
        val parts = rawDate.split("-")
        if (parts.size == 3) "${parts[1].toInt()}/${parts[2].toInt()}/${parts[0]}" else rawDate
    } catch (e: Exception) { rawDate }
    
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Col 1: User (weight 2.5f)
        Row(
            modifier = Modifier.weight(2.5f),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Surface(
                modifier = Modifier.size(32.dp),
                shape = RoundedCornerShape(16.dp),
                color = NetflixGray,
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(initial, color = NetflixWhite, fontWeight = FontWeight.Medium, fontSize = 14.sp)
                }
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column {
                Text(displayName, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                Text(user.username, color = NetflixLightGray, fontSize = 12.sp)
            }
        }
        
        // Col 2: Role (weight 1.2f)
        Box(modifier = Modifier.weight(1.2f), contentAlignment = Alignment.CenterStart) {
            val roleColor = if (user.isAdmin) Color(0xFFA855F7) else Color(0xFF3B82F6)
            val roleBg = if (user.isAdmin) Color(0x33A855F7) else Color(0x333B82F6)
            Surface(color = roleBg, shape = RoundedCornerShape(4.dp)) {
                Text(
                    text = if (user.isAdmin) "Admin" else "User", 
                    color = roleColor, 
                    fontSize = 11.sp, 
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp), 
                    fontWeight = FontWeight.Medium
                )
            }
        }
        
        // Col 3: Created (weight 1.5f)
        Text(fmtDate, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.weight(1.5f))
        
        // Col 4: Actions (weight 2f)
        Row(
            modifier = Modifier.weight(2f), 
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Button(
                onClick = onEdit,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
                shape = RoundedCornerShape(4.dp),
                modifier = Modifier.height(28.dp).defaultMinSize(minWidth = 1.dp)
            ) {
                Text("Edit", fontSize = 11.sp, color = NetflixWhite)
            }
            
            if (canDelete) {
                Button(
                    onClick = onDelete,
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0x33EF4444)), // Red 400 / 20%
                    contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
                    shape = RoundedCornerShape(4.dp),
                    modifier = Modifier.height(28.dp).defaultMinSize(minWidth = 1.dp)
                ) {
                    Text("Delete", fontSize = 11.sp, color = Color(0xFFF87171)) // Red 400
                }
            }
        }
    }
}

@Composable
private fun ActivitySection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Activity Feed", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)

        if (uiState.isLoading && uiState.activityLogs.isEmpty()) {
            Text("Loading activity...", color = NetflixLightGray)
        } else if (uiState.activityLogs.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text("No activity found", color = NetflixLightGray, fontSize = 14.sp)
                }
            }
        } else {
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column {
                    // Header Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Time", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.8f))
                        Text("User", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.2f))
                        Text("Action", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                        Text("Media", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2.5f))
                    }
                    
                    HorizontalDivider(color = Color.White.copy(alpha = 0.05f))

                    // Activity List
                    uiState.activityLogs.forEachIndexed { index, log ->
                        if (index > 0) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.02f))
                        }
                        ActivityTableRow(log)
                    }
                }
            }
        }
    }
}

@Composable
private fun ActivityTableRow(log: com.velox.app.data.model.ActivityLogDto) {
    val fmtDate = try {
        // e.g. "2026-04-10T11:22:25Z" -> "4/10/2026, 11:22:25"
        val parts = log.createdAt.split("T")
        val dateParts = parts[0].split("-")
        val yyyy = dateParts[0]
        val mm = dateParts[1].toInt()
        val dd = dateParts[2].toInt()
        val timePart = parts[1].replace("Z", "")
        "$mm/$dd/$yyyy, $timePart"
    } catch (e: Exception) {
        log.createdAt
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Time
        Text(fmtDate, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.weight(1.8f))
        
        // User
        Text(log.username ?: "System", color = NetflixWhite, fontSize = 14.sp, modifier = Modifier.weight(1.2f))
        
        // Action
        Box(modifier = Modifier.weight(1.5f), contentAlignment = Alignment.CenterStart) {
            ActionBadge(log.action)
        }
        
        // Media
        val targetStr = log.mediaTitle?.takeIf { it.isNotBlank() } ?: "-"
        Text(targetStr, color = NetflixLightGray, fontSize = 13.sp, maxLines = 1, overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis, modifier = Modifier.weight(2.5f))
    }
}

@Composable
private fun ActionBadge(action: String) {
    val (bgColor, textColor) = when (action) {
        "login" -> Pair(Color(0x333B82F6), Color(0xFF60A5FA)) // blue 500/20, blue 400
        "play_start" -> Pair(Color(0x3322C55E), Color(0xFF4ADE80)) // green 500/20, green 400
        "play_stop" -> Pair(Color(0x33EAB308), Color(0xFFFACC15)) // yellow 500/20, yellow 400
        "library_scan" -> Pair(Color(0x33A855F7), Color(0xFFC084FC)) // purple 500/20, purple 400
        "media_added" -> Pair(Color(0x3314B8A6), Color(0xFF2DD4BF)) // teal 500/20, teal 400
        else -> Pair(Color(0x336B7280), Color(0xFF9CA3AF)) // gray
    }

    Surface(
        color = bgColor,
        shape = RoundedCornerShape(4.dp)
    ) {
        Text(
            text = action,
            color = textColor,
            fontSize = 10.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
        )
    }
}

@Composable
private fun TasksSection(viewModel: SettingsViewModel, uiState: SettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Scheduled Tasks", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)

        if (uiState.tasks.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text("No tasks configured", color = NetflixLightGray, fontSize = 14.sp)
                }
            }
        } else {
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column {
                    // Header Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Task", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
                        Text("Interval", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1f))
                        Text("Last / Next Run", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
                        Text("Status", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1f))
                        Text("Action", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                    }
                    
                    HorizontalDivider(color = Color.White.copy(alpha = 0.05f))

                    // Tasks List
                    val sortedTasks = uiState.tasks.sortedBy { it.name }
                    sortedTasks.forEachIndexed { index, task ->
                        if (index > 0) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.02f))
                        }
                        TaskTableRow(
                            task = task, 
                            runningTaskName = uiState.runningTaskName, 
                            onRun = { viewModel.runTask(task.name) }
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun TaskTableRow(
    task: com.velox.app.data.model.TaskDto,
    runningTaskName: String?,
    onRun: () -> Unit
) {
    val isRunning = task.running || runningTaskName == task.name
    
    fun formatDateTime(dt: String?): String {
        if (dt == null) return "Never"
        return try {
            val parts = dt.split("T")
            val dateParts = parts[0].split("-")
            val yyyy = dateParts[0]
            val mm = dateParts[1].toInt()
            val dd = dateParts[2].toInt()
            val timePart = parts[1].replace("Z", "").take(5)
            "$mm/$dd/$yyyy, $timePart"
        } catch (e: Exception) {
            dt
        }
    }

    val lastRun = formatDateTime(task.lastRun)
    val nextRun = formatDateTime(task.nextRun)

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Task
        Text(task.name, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
        
        // Interval
        Text(task.interval, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.weight(1f))
        
        // Timestamps
        Column(modifier = Modifier.weight(2f)) {
            Text("Last: $lastRun", color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(2.dp))
            Text("Next: $nextRun", color = NetflixLightGray, fontSize = 12.sp)
        }
        
        // Status
        Box(modifier = Modifier.weight(1f), contentAlignment = Alignment.CenterStart) {
            if (isRunning) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    CircularProgressIndicator(modifier = Modifier.size(12.dp), color = Color(0xFFFACC15), strokeWidth = 1.5.dp)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text("Running", color = Color(0xFFFACC15), fontSize = 12.sp)
                }
            } else {
                Text("Idle", color = Color.Gray, fontSize = 12.sp)
            }
        }
        
        // Action
        Box(modifier = Modifier.weight(1.5f), contentAlignment = Alignment.CenterStart) {
            Button(
                onClick = onRun,
                enabled = !isRunning,
                colors = ButtonDefaults.buttonColors(
                    containerColor = NetflixGray,
                    disabledContainerColor = NetflixGray.copy(alpha = 0.5f)
                ),
                contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
                shape = RoundedCornerShape(4.dp),
                modifier = Modifier.height(28.dp).defaultMinSize(minWidth = 1.dp)
            ) {
                Icon(if (isRunning) LucideIcons.Refresh else LucideIcons.PlayCircle, contentDescription = null, modifier = Modifier.size(12.dp), tint = NetflixWhite)
                Spacer(modifier = Modifier.width(4.dp))
                Text(if (isRunning) "Running..." else "Run Now", fontSize = 11.sp, color = NetflixWhite)
            }
        }
    }
}

@Composable
private fun WebhooksSection(
    viewModel: SettingsViewModel?,
    uiState: SettingsUiState,
    onAddClick: () -> Unit,
    onEditClick: (Webhook) -> Unit,
    onDeleteClick: (Int) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text("Webhooks", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Button(
                onClick = onAddClick,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Add, contentDescription = null)
                Spacer(modifier = Modifier.width(4.dp))
                Text("Add Webhook")
            }
        }

        if (uiState.webhooks.isEmpty()) {
            Text("No webhooks configured", color = NetflixLightGray)
        } else {
            uiState.webhooks.forEach { webhook ->
                WebhookCard(
                    webhook = webhook,
                    onEdit = { onEditClick(webhook) },
                    onDelete = { onDeleteClick(webhook.id) },
                )
            }
        }
    }
}

@Composable
private fun WebhookCard(
    webhook: Webhook,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
) {
    var showMenu by remember { mutableStateOf(false) }

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(webhook.url, color = NetflixWhite)
                Text(webhook.events.joinToString(", "), color = NetflixLightGray, fontSize = 12.sp)
            }
            Surface(
                shape = RoundedCornerShape(4.dp),
                color = if (webhook.active) NetflixRed.copy(alpha = 0.2f) else NetflixGray,
            ) {
                Text(
                    if (webhook.active) "Active" else "Inactive",
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                    color = if (webhook.active) NetflixRed else NetflixLightGray,
                    fontSize = 10.sp,
                )
            }
            Box {
                IconButton(onClick = { showMenu = true }) {
                    Icon(LucideIcons.MoreVert, contentDescription = "More", tint = NetflixLightGray)
                }
                DropdownMenu(
                    expanded = showMenu,
                    onDismissRequest = { showMenu = false },
                ) {
                    DropdownMenuItem(
                        text = { Text("Edit") },
                        onClick = { onEdit(); showMenu = false },
                        leadingIcon = { Icon(LucideIcons.Edit, contentDescription = null) },
                    )
                    DropdownMenuItem(
                        text = { Text("Delete", color = NetflixRed) },
                        onClick = { onDelete(); showMenu = false },
                        leadingIcon = { Icon(LucideIcons.Delete, contentDescription = null, tint = NetflixRed) },
                    )
                }
            }
        }
    }
}

// Dialogs

@Composable
private fun LibraryDialog(
    library: Library?,
    onDismiss: () -> Unit,
    onSave: (name: String, type: String, paths: List<String>) -> Unit,
) {
    var name by remember { mutableStateOf(library?.name ?: "") }
    var type by remember { mutableStateOf(library?.type ?: "movie") }
    var paths by remember { mutableStateOf(library?.paths?.joinToString(",") ?: "") }
    var typeExpanded by remember { mutableStateOf(false) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (library != null) "Edit Library" else "Add Library", color = NetflixWhite) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text("Name") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
                Box {
                    OutlinedTextField(
                        value = type.replaceFirstChar { it.uppercase() },
                        onValueChange = {},
                        label = { Text("Type") },
                        modifier = Modifier.fillMaxWidth(),
                        colors = textFieldColors(),
                        readOnly = true,
                        trailingIcon = {
                            IconButton(onClick = { typeExpanded = true }) {
                                Icon(LucideIcons.ArrowDropDown, contentDescription = null)
                            }
                        },
                    )
                    DropdownMenu(
                        expanded = typeExpanded,
                        onDismissRequest = { typeExpanded = false },
                    ) {
                        listOf("movie", "series").forEach { t ->
                            DropdownMenuItem(
                                text = { Text(t.replaceFirstChar { it.uppercase() }) },
                                onClick = { type = t; typeExpanded = false },
                            )
                        }
                    }
                }
                OutlinedTextField(
                    value = paths,
                    onValueChange = { paths = it },
                    label = { Text("Paths (comma separated)") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (name.isNotBlank() && paths.isNotBlank()) {
                        onSave(name, type, paths.split(",").map { it.trim() }.filter { it.isNotEmpty() })
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            ) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel", color = NetflixLightGray)
            }
        },
        containerColor = NetflixDark,
    )
}

@Composable
private fun UserDialog(
    user: AdminUser?,
    onDismiss: () -> Unit,
    onSave: (username: String, password: String, displayName: String, isAdmin: Boolean) -> Unit,
) {
    var username by remember { mutableStateOf(user?.username ?: "") }
    var password by remember { mutableStateOf("") }
    var displayName by remember { mutableStateOf(user?.displayName ?: "") }
    var isAdmin by remember { mutableStateOf(user?.isAdmin ?: false) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (user != null) "Edit User" else "Add User", color = NetflixWhite) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = username,
                    onValueChange = { username = it },
                    label = { Text("Username") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                    enabled = user == null, // Can't change username when editing
                )
                if (user == null) {
                    OutlinedTextField(
                        value = password,
                        onValueChange = { password = it },
                        label = { Text("Password") },
                        modifier = Modifier.fillMaxWidth(),
                        colors = textFieldColors(),
                    )
                }
                OutlinedTextField(
                    value = displayName,
                    onValueChange = { displayName = it },
                    label = { Text("Display Name") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("Admin", color = NetflixWhite)
                    Switch(
                        checked = isAdmin,
                        onCheckedChange = { isAdmin = it },
                        colors = SwitchDefaults.colors(checkedThumbColor = NetflixRed, checkedTrackColor = NetflixRed),
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (username.isNotBlank() && displayName.isNotBlank() && (user != null || password.isNotBlank())) {
                        onSave(username, password, displayName, isAdmin)
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            ) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel", color = NetflixLightGray)
            }
        },
        containerColor = NetflixDark,
    )
}

@Composable
private fun WebhookDialog(
    webhook: Webhook?,
    onDismiss: () -> Unit,
    onSave: (url: String, events: List<String>, active: Boolean) -> Unit,
) {
    var url by remember { mutableStateOf(webhook?.url ?: "") }
    var eventsText by remember { mutableStateOf(webhook?.events?.joinToString(", ") ?: "") }
    var active by remember { mutableStateOf(webhook?.active ?: true) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (webhook != null) "Edit Webhook" else "Add Webhook", color = NetflixWhite) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = url,
                    onValueChange = { url = it },
                    label = { Text("URL") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
                OutlinedTextField(
                    value = eventsText,
                    onValueChange = { eventsText = it },
                    label = { Text("Events (comma separated)") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                    placeholder = { Text("media_added, library_scan_complete", color = NetflixGray) },
                )
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("Active", color = NetflixWhite)
                    Switch(
                        checked = active,
                        onCheckedChange = { active = it },
                        colors = SwitchDefaults.colors(checkedThumbColor = NetflixRed, checkedTrackColor = NetflixRed),
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (url.isNotBlank()) {
                        onSave(
                            url,
                            eventsText.split(",").map { it.trim() }.filter { it.isNotEmpty() },
                            active,
                        )
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            ) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel", color = NetflixLightGray)
            }
        },
        containerColor = NetflixDark,
    )
}

// Helper components

@Composable
private fun SettingsToggle(
    label: String,
    description: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(label, color = NetflixWhite)
                Text(description, color = NetflixLightGray, fontSize = 12.sp)
            }
            Switch(
                checked = checked,
                onCheckedChange = onCheckedChange,
                colors = SwitchDefaults.colors(checkedThumbColor = NetflixRed, checkedTrackColor = NetflixRed),
            )
        }
    }
}

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
private fun SettingsDropdown(
    label: String,
    value: String,
    options: List<Pair<String, String>>,
    onValueChange: (String) -> Unit,
) {
    var expanded by remember { mutableStateOf(false) }
    val displayValue = options.find { it.first == value }?.second ?: value

    Column {
        Text(label, color = NetflixWhite, fontSize = 14.sp)
        Spacer(modifier = Modifier.height(4.dp))
        Box(modifier = Modifier.fillMaxWidth()) {
            OutlinedTextField(
                value = displayValue,
                onValueChange = {},
                readOnly = true,
                modifier = Modifier.fillMaxWidth(),
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                trailingIcon = {
                    Icon(LucideIcons.ArrowDropDown, contentDescription = null, tint = NetflixWhite)
                },
            )
            // Transparent touch interceptor overlay
            Surface(
                modifier = Modifier
                    .matchParentSize()
                    .clickable { expanded = true },
                color = androidx.compose.ui.graphics.Color.Transparent
            ) {}
        }

        if (expanded) {
            ModalBottomSheet(
                onDismissRequest = { expanded = false },
                containerColor = NetflixDark,
                dragHandle = { BottomSheetDefaults.DragHandle() }
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 32.dp)
                ) {
                    Text(
                        text = "Select $label",
                        color = NetflixLightGray,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(horizontal = 24.dp, vertical = 8.dp)
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    options.forEach { option ->
                        val isSelected = option.first == value
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    onValueChange(option.first)
                                    expanded = false
                                }
                                .padding(horizontal = 24.dp, vertical = 16.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Text(
                                text = option.second,
                                color = if (isSelected) NetflixWhite else NetflixLightGray,
                                fontSize = 16.sp,
                                fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal
                            )
                            if (isSelected) {
                                Icon(
                                    imageVector = LucideIcons.Check,
                                    contentDescription = "Selected",
                                    tint = NetflixRed,
                                    modifier = Modifier.size(20.dp)
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun textFieldColors() = OutlinedTextFieldDefaults.colors(
    focusedTextColor = NetflixWhite,
    unfocusedTextColor = NetflixWhite,
    focusedBorderColor = NetflixRed,
    unfocusedBorderColor = NetflixGray,
    focusedLabelColor = NetflixRed,
    unfocusedLabelColor = NetflixLightGray,
)

@Preview(showBackground = true)
@Composable
fun SettingsScreenPreview() {
    VeloxTheme {
        SettingsContent(
            uiState = SettingsUiState(),
            onBackClick = {},
            onProfileClick = {},
            onNotificationsClick = {},
            onAction = {},
        )
    }
}

@Composable
private fun SettingsGroupCard(
    tabs: List<SettingsTab>,
    onAction: (SettingsAction) -> Unit
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        color = NetflixDark,
        shape = RoundedCornerShape(16.dp)
    ) {
        Column {
            tabs.forEachIndexed { index, tab ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onAction(SettingsAction.SelectTab(tab.name.lowercase())) }
                        .padding(horizontal = 16.dp, vertical = 16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = iconForTab(tab),
                        contentDescription = null,
                        tint = NetflixLightGray,
                        modifier = Modifier.size(22.dp)
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Text(tab.title, color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.Medium)
                    Spacer(Modifier.weight(1f))
                    Icon(LucideIcons.ChevronRight, contentDescription = null, tint = NetflixGray, modifier = Modifier.size(20.dp))
                }
                
                if (index < tabs.size - 1) {
                    androidx.compose.material3.HorizontalDivider(
                        color = NetflixBlack,
                        modifier = Modifier.padding(start = 54.dp)
                    )
                }
            }
        }
    }
}

private fun iconForTab(tab: SettingsTab): androidx.compose.ui.graphics.vector.ImageVector {
    return when (tab) {
        SettingsTab.PROFILE -> LucideIcons.Person
        SettingsTab.PREFERENCES -> LucideIcons.Settings
        SettingsTab.SECURITY -> LucideIcons.Lock
        SettingsTab.SESSIONS -> LucideIcons.Devices
        SettingsTab.METADATA -> LucideIcons.Info
        SettingsTab.SUBTITLES -> LucideIcons.Subtitles
        SettingsTab.PLAYBACK -> LucideIcons.PlayCircle
        SettingsTab.CINEMA -> LucideIcons.Movie
        SettingsTab.PRETRANSCODE -> LucideIcons.FlashOn
        SettingsTab.MARKERS -> LucideIcons.SkipNext
        SettingsTab.DASHBOARD -> LucideIcons.GridView
        SettingsTab.LIBRARIES -> LucideIcons.Folder
        SettingsTab.USERS -> LucideIcons.Person
        SettingsTab.ACTIVITY -> LucideIcons.ShowChart
        SettingsTab.TASKS -> LucideIcons.CheckCircle
        SettingsTab.WEBHOOKS -> LucideIcons.Link
        else -> LucideIcons.Settings
    }
}

@Composable
fun FeedbackCard() {
    val context = LocalContext.current
    val appVersion = com.velox.app.BuildConfig.VERSION_NAME
    val vecode = com.velox.app.BuildConfig.VERSION_CODE
    val device = android.os.Build.MODEL
    val osVersion = android.os.Build.VERSION.RELEASE
    val apiLevel = android.os.Build.VERSION.SDK_INT

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 24.dp, vertical = 8.dp)
    ) {
        Text(
            text = "SUPPORT",
            color = NetflixLightGray,
            fontSize = 12.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .clickable {
                    val intent = Intent(Intent.ACTION_SENDTO).apply {
                        data = Uri.parse("mailto:thanglong2098@gmail.com")
                        putExtra(Intent.EXTRA_SUBJECT, "[Velox Bug Report] Feedback from Android App")
                        putExtra(
                            Intent.EXTRA_TEXT,
                            "Vui lòng mô tả lỗi hoặc góp ý của bạn ở dưới đây:\n\n\n\n\n\n" +
                            "========================\n" +
                            "App Version: \$appVersion (\$vecode)\n" +
                            "Device Model: \$device\n" +
                            "Android Version: \$osVersion (API \$apiLevel)\n" +
                            "========================\n"
                        )
                    }
                    try {
                        context.startActivity(intent)
                    } catch (e: Exception) {
                        android.widget.Toast.makeText(context, "No email app installed.", android.widget.Toast.LENGTH_SHORT).show()
                    }
                },
            colors = CardDefaults.cardColors(containerColor = Color(0xFF1E1E1E)),
            shape = RoundedCornerShape(12.dp)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.Default.Email,
                    contentDescription = "Report Bug",
                    tint = NetflixWhite,
                    modifier = Modifier.size(24.dp)
                )
                Spacer(modifier = Modifier.width(16.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text("Report a Bug / Feedback", color = NetflixWhite, fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
                    Text("Gửi email trực tiếp cho tác giả", color = NetflixLightGray, fontSize = 13.sp)
                }
                Icon(
                    imageVector = LucideIcons.ChevronRight,
                    contentDescription = null,
                    tint = NetflixGray,
                    modifier = Modifier.size(20.dp)
                )
            }
        }
    }
}

@Composable
fun AppVersionCard() {
    val context = LocalContext.current
    val coroutineScope = rememberCoroutineScope()
    var latestVersion by remember { mutableStateOf<String?>(null) }
    var downloadUrl by remember { mutableStateOf<String?>(null) }
    var isChecking by remember { mutableStateOf(false) }
    var updateStatus by remember { mutableStateOf("") }
    var isMandatory by remember { mutableStateOf(false) }
    
    val currentVersion = com.velox.app.BuildConfig.VERSION_NAME
    val currentVersionCode = com.velox.app.BuildConfig.VERSION_CODE

    val checkUpdate = {
        isChecking = true
        updateStatus = "Checking..."
        coroutineScope.launch(Dispatchers.IO) {
            try {
                // Remove trailing slash if present, then append API path
                val baseUrl = com.velox.app.BuildConfig.BASE_URL.trimEnd('/')
                val url = URL("$baseUrl/app-versions/latest?platform=android")
                val connection = url.openConnection() as HttpURLConnection
                connection.requestMethod = "GET"
                connection.setRequestProperty("Accept", "application/json")
                
                if (connection.responseCode == 200) {
                    val response = connection.inputStream.bufferedReader().readText()
                    val json = JSONObject(response)
                    val vName = json.getString("version_name")
                    val vCode = json.getInt("version_code")
                    val dUrl = json.getString("download_url")
                    val isMandatoryFlag = json.optBoolean("is_mandatory", false)
                    
                    withContext(Dispatchers.Main) {
                        isChecking = false
                        if (vCode > currentVersionCode) {
                            latestVersion = vName
                            downloadUrl = dUrl
                            isMandatory = isMandatoryFlag
                            
                            updateStatus = if (isMandatoryFlag) {
                                "Breaking change! Mandatory update required: $vName"
                            } else {
                                "New version available: $vName"
                            }
                        } else {
                            updateStatus = "App is up to date"
                        }
                    }
                } else if (connection.responseCode == 404) {
                    withContext(Dispatchers.Main) {
                        isChecking = false
                        updateStatus = "App is up to date"
                    }
                } else {
                    withContext(Dispatchers.Main) {
                        isChecking = false
                        updateStatus = "Update check failed (Code: ${connection.responseCode})"
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    isChecking = false
                    updateStatus = "Error checking update: ${e.message}"
                }
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 24.dp, vertical = 24.dp)
    ) {
        Text(
            text = "APP VERSION",
            color = NetflixLightGray,
            fontSize = 12.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = Color(0xFF1E1E1E)),
            shape = RoundedCornerShape(12.dp) // Match SettingsGroupCard
        ) {
            Column(
                modifier = Modifier.fillMaxWidth().padding(16.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Image(
                        painter = painterResource(id = com.velox.app.R.drawable.velox_logo),
                        contentDescription = "App Logo",
                        modifier = Modifier.size(48.dp).clip(RoundedCornerShape(8.dp))
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Velox for Android", color = NetflixWhite, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                        Text("Version $currentVersion", color = NetflixLightGray, fontSize = 13.sp)
                    }
                    
                    // Refresh Button
                    if (latestVersion == null) {
                        IconButton(
                            onClick = { if (!isChecking) checkUpdate() },
                            modifier = Modifier.background(NetflixGray.copy(alpha = 0.2f), RoundedCornerShape(50))
                        ) {
                            if (isChecking) {
                                CircularProgressIndicator(modifier = Modifier.size(20.dp), color = NetflixWhite, strokeWidth = 2.dp)
                            } else {
                                Icon(LucideIcons.Refresh, contentDescription = "Check for Updates", tint = NetflixWhite, modifier = Modifier.size(20.dp))
                            }
                        }
                    }
                }
                
                if (updateStatus.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(16.dp))
                    HorizontalDivider(color = NetflixGray.copy(alpha = 0.2f))
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    if (latestVersion != null && downloadUrl != null) {
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Column(modifier = Modifier.weight(1f)) {
                                Text("Update Available", color = NetflixWhite, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                                Text(updateStatus, color = if (isMandatory) NetflixRed else NetflixLightGray, fontSize = 13.sp)
                            }
                            Spacer(modifier = Modifier.width(12.dp))
                            Button(
                                onClick = {
                                    val intent = Intent(Intent.ACTION_VIEW, Uri.parse(downloadUrl))
                                    context.startActivity(intent)
                                },
                                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp)
                            ) {
                                Text(if (isMandatory) "UPDATE" else "DOWNLOAD", fontWeight = FontWeight.Bold)
                            }
                        }
                    } else {
                        Text(updateStatus, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.fillMaxWidth(), textAlign = TextAlign.Center)
                    }
                }
            }
        }
    }
}
